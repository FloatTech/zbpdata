#!/bin/bash
approot=$(pwd)
appname={{.Appname}}
appexec={{.Command}}
isdaemon=${isdaemon:=true}
usemonitor=${usemonitor:=false}

##########################################
# main function.
##########################################
function main() {
    echo "=="$(date  +"%Y-%m-%d %H:%M:%S")"=="  
    case "$1" in
        start)
            $usemonitor && start_by_monitor     || start_directly
            ;;
        restart|reload)
            $usemonitor && restart_by_monitor   || restart_directly
            ;;
        stop)
            $usemonitor && stop_by_monitor      || stop_directly
            ;;
        install)
            $usemonitor && install_with_monitor || install_no_monitor
            ;;
        *)
        echo "Usage: ./load.sh {install|start|stop|restart}"
        exit 1
    esac
}

##########################################
# install: 
#  do something before reload after deploy 
##########################################
function install_no_monitor() {
    ##### make logs directory at app directory ##############
    function isemptydir() {
        [ `ls -A "$approot/logs" 2>/dev/null | wc -l` -eq 0 ] && echo true || echo false;
    }
    if [ ! -d $approot/logs -o $(isemptydir) == true ]; then
        mkdir -p /data/logs/${appname};
        rm -rf $approot/logs;
        ln -s /data/logs/${appname} $approot/logs;
    fi
}
function install_with_monitor() {
    ##### do install no monitor actions #####################
    install_no_monitor

    ##### get the supervisor config & incloude directory ####
    appmonitor=supervisor
    monitconf=${SUPERVISOR_CONF:=$(find /etc -name supervisord.conf -type f 2>/dev/null)}
    monitconf=${monitconf:=/etc/supervisor/supervisord.conf}
    echo "The path of supervisord.conf is: '${monitconf}'"
    monitinclude=${SUPERVISOR_INCLUDE:=$(grep '^\[include\]$' $monitconf -A2 | sed -n 's/^files=\([^*]*\)\(\*.conf\)*/\1/p')}
    monitinclude=${monitinclude:=/etc/supervisord.d}
    echo "The folder of supervisord files is: ${monitincloude}"

    ##### install ignore if file exist. #####################
    if [ -f ${monitinclude}/${appname}.ini ]; then
       echo "WARNING: install ignore, because of file '${monitinclude}/${appname}.ini' is exist"
       exit 0
    fi
  
    cat >${monitinclude}/${appname}.ini <<-EOF
	[program:${appname}]
	command=${appexec}
	process_name=%(program_name)s
	numprocs=1
	priority=1
	directory=${approot}
	autostart=true
	autorestart=true

	stdout_logfile=/data/logs/supervisor/%(program_name)s.log
	stderr_logfile=/data/logs/supervisor/%(program_name)s.err
	
	stdout_logfile_maxbytes=500MB
	stderr_logfile_maxbytes=500MB
	stdout_logfile_backups=5
	stderr_logfile_backups=5
EOF


    ##### update supervisor & show supervisor status ########
    echo "Update ing..."
    supervisorctl update
    echo "Update done"
    echo "-----current supervisor result: -----";
    supervisorctl status ${appname}
    echo "-----current supervisor result done  -----";
    
}

##########################################
# start: 
#   start the application 
##########################################
function start_directly() {
    if $isdaemon; then
        #### start app backgroud #####
        nohup $appexec  &
        echo $! > $approot/logs/pid
        wait $! 
    else
        #### start app frontgroud ####
        rm -f $approot/logs/stoped
        while [ ! -f $approot/logs/stoped ]; do
          $appexec 
          sleep 30 
        done
    fi
}

function start_by_monitor() {
    echo "Start application ...";
    supervisorctl start ${appname} 
    sleep 0.2;
    echo "-----current supervisor result: -----";
    supervisorctl status ${appname}
    echo "-----current supervisor result done  -----";
    echo "START DONE";
}

##########################################
# stop: 
#   stop the application 
##########################################
function stop_directly() {
    #### check pid file ############
    if [ ! -f $approot/logs/pid ]; then
      echo "ERROR: no pid file: '$approot/logs/pid'"
      return 1 
    fi
    #### check cmdline  ############
    local pid=$(cat $approot/logs/pid)
    if [ -z $pid ] || [ "$(/proc/$pid/cmdline)"x == "$appexec"x ]; then
      echo "ERROR: pid($pid) is invalid or command line mismatch";
      return 1 
    fi
    #### do stop  ##################
    echo "stoped at: $(date  +'%Y-%m-%d %H:%M:%S')" > $approot/logs/stoped
    kill -9 $pid
}
function stop_by_monitor() {
    echo "Stop application ...";
    supervisorctl stop ${appname}
    sleep 0.1;
    echo "-----current supervisor result: -----";
    supervisorctl status ${appname}
    echo "-----current supervisor result done  -----";
    echo "STOP DONE"
}

##########################################
# restart: 
#   restart the application 
##########################################
function restart_directly() {
    stop_directly && start_directly
}
function restart_by_monitor() {
    stop_by_monitor && start_by_monitor
}

#############################################
############## run ##########################
main $@
#############################################
