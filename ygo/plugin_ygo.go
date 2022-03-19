//本插件查卡通过网页"https://www.ygo-sem.cn/"获取的
package plugin_ygo

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/web"
)

var (
	reqconf = [...]string{"GET", "https://www.ygo-sem.cn/",
		"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36"}
	searchName string
	searchpage int
	pagemax    int
	CardsName  = []string{}
	CardsHref  = []string{}
)

func RegexpMatch(rule string, str string) (regexpresult string, err bool) {
	regexp_rule := regexp.MustCompile(rule)
	regexp_result := regexp_rule.FindAllStringSubmatch(str, -1)
	if len(regexp_result) == 0 {
		err = true
		return
	}
	regexpresult = strings.Replace(regexp_result[0][1], "\r\n", "", -1) //去掉空格
	regexpresult = strings.Replace(regexpresult, " ", "", -1)           //去掉空格
	return regexpresult, false
}

func init() {
	en := control.Register("ygo", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "1.指令：ygo XXX\n" +
			"①xxx为卡名：\n查询卡名为XXX的卡信息\n" +
			"②xxx为“随机一卡”：\n随机展示一张卡\n" +
			"2.指令：x\n①x为搜索列表对应的数字\n获取对应的卡片信息\n②x为“下一页”：\n搜索列表翻到下一页",
	})

	en.OnRegex(`^ygo (.+)$`, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		List := ctx.State["regex_matched"].([]string)[1]
		switch List {
		case "随机一卡":
			url := "https://www.ygo-sem.cn/Cards/Default.aspx"
			// 请求html页面
			list_body, err := web.ReqWith(url, reqconf[0], reqconf[1], reqconf[2])
			if err != nil {
				fmt.Println("网页数据读取错误：", err)
			}
			//获取卡牌数量
			listmax, regexpResult := RegexpMatch("条 共:(?s:(.*?))条</span>", string(list_body))
			if regexpResult {
				fmt.Println("数据存在错误: 无法获取当前卡池数量")
			}
			Listmax, _ := strconv.Atoi(listmax) //fmt.Println("当前总卡数：", listmax)
			List = "q=" + fmt.Sprint(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(Listmax))
		default:
			searchName = List
			List = "q=" + searchName
		}
		url := "https://www.ygo-sem.cn/Cards/S.aspx?" + List
		// 请求html页面
		body, err := web.ReqWith(url, reqconf[0], reqconf[1], reqconf[2])
		if err != nil {
			fmt.Println("网页数据读取错误：", err)
		}
		//获取卡牌数量
		listmax, regexpResult := RegexpMatch("找到(?s:(.*?))个卡片", string(body))
		switch regexpResult {
		case true:
			card_data, imageBase64 := getYGOdata(string(body))
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Image("base64://"+imageBase64), message.Text(card_data)))
		case false:
			//清空内容
			searchpage = 0
			CardsName = []string{}
			CardsHref = []string{}
			//筛选数据
			list_data := getYGolist(listmax, string(body))
			ctx.SendChain(message.Text(list_data))
			if len(CardsName) != 0 {
				//等待用户下一步选择
				var next = zero.NewFutureEvent("message", 999, false, zero.RegexRule(`(下一页)|\d+`), zero.OnlyGroup, zero.CheckUser(ctx.Event.UserID))
				for {
					select {
					case <-time.After(time.Second * 120):
						ctx.Send(
							message.ReplyWithMessage(ctx.Event.MessageID,
								message.Text("等待超时,搜索结束"),
							),
						)
						return
					case e := <-next.Next():
						List := e.Message.String()
						switch List {
						case "下一页":
							searchpage += 1
							if searchpage > pagemax {
								searchpage = 0
								ctx.SendChain(message.At(ctx.Event.UserID), message.Text("已是最后一页，返回到第一页"))
							}
							url := "https://www.ygo-sem.cn/Cards/S.aspx?dRace=&attr=&q=" + searchName + "&start=" + strconv.Itoa(searchpage*30)
							// 请求html页面
							body, err := web.ReqWith(url, reqconf[0], reqconf[1], reqconf[2])
							if err != nil {
								fmt.Println("网页数据读取错误：", err)
							}
							//获取卡牌数量
							listmax, _ := RegexpMatch("找到(?s:(.*?))个卡片", string(body))
							//清空内容
							CardsName = []string{}
							CardsHref = []string{}
							//筛选数据
							list_data := getYGolist(listmax, string(body))
							ctx.SendChain(message.Text(list_data))
						default:
							Cardint, _ := strconv.Atoi(List)
							if Cardint < len(CardsName) {
								url := "https://www.ygo-sem.cn/" + CardsHref[Cardint]
								// 请求html页面
								body, err := web.ReqWith(url, reqconf[0], reqconf[1], reqconf[2])
								if err != nil {
									fmt.Println("网页数据读取错误：", err)
								}
								card_data, imageBase64 := getYGOdata(string(body))
								ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Image("base64://"+imageBase64), message.Text(card_data)))
								return
							} else {
								ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入正确的序号"))
							}
						}
					}
				}
			}
		}
	})

	//每天12点随机分享一张卡
	_, err := process.CronTab.AddFunc("00 12 * * *", func() {
		m, ok := control.Lookup("ygo")
		if !ok {
			return
		}
		url := "https://www.ygo-sem.cn/Cards/Default.aspx"
		// 请求html页面
		list_body, err := web.ReqWith(url, reqconf[0], reqconf[1], reqconf[2])
		if err != nil {
			// 错误处理
			fmt.Println("网页数据读取错误：", err)
		}
		//fmt.Println(string(list_body))
		result_number := regexp.MustCompile("条 共:(?s:(.*?))条</span>")
		list_number := result_number.FindAllStringSubmatch(string(list_body), -1)
		if len(list_number) == 0 {
			fmt.Println("数据存在错误: 无法获取当前卡池数量")
		}
		listmax, _ := strconv.Atoi(list_number[0][1])
		//fmt.Println("当前总卡数：", listmax)
		List := "q=" + fmt.Sprint(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(listmax))
		url = "https://www.ygo-sem.cn/Cards/S.aspx?" + List
		// 请求html页面
		body, err := web.ReqWith(url, reqconf[0], reqconf[1], reqconf[2])
		if err != nil {
			// 错误处理
			fmt.Println("网页数据读取错误：", err)
		}
		//fmt.Println(string(body))
		//筛选数据
		card_data, imageBase64 := getYGOdata(string(body))
		zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
			for _, g := range ctx.GetGroupList().Array() {
				grp := g.Get("group_id").Int()
				if m.IsEnabledIn(grp) {
					//输出数据
					ctx.SendGroupMessage(grp, message.Message{message.Text("当前游戏王卡池总数：" + list_number[0][1] + "\n\n今日分享卡牌：\n\n"), message.Image("base64://" + imageBase64), message.Text(card_data)})
					process.SleepAbout1sTo2s()
				}
			}
			return true
		})
	})
	if err != nil {
		fmt.Println(err)
	} //*/
}

//获取单卡信息
func getYGOdata(body string) (ygodata string, imageBase64 string) {
	//获取卡图连接
	cardpic, regexpResult := RegexpMatch("picsCN(?s:(.*?)).jpg", body)
	if regexpResult {
		return "数据存在错误: 无法获取卡图", ""
	}
	pic_href := "https://www.ygo-sem.cn/yugioh/picsCN" + cardpic + ".jpg"
	// 读取获取的[]byte数据
	data, _ := web.ReqWith(pic_href, reqconf[0], reqconf[1], reqconf[2])
	imageBase64 = base64.StdEncoding.EncodeToString(data)

	//获取卡名*/
	card_name, regexpResult := RegexpMatch("<b>中文名</b> </span>&nbsp;<span class=\"item_box_value\">(?s:(.*?))</span>", body)
	if regexpResult {
		return "数据存在错误: 无法获取卡名", ""
	}
	ygodata = ygodata + card_name + "\n"

	//获取属性
	/*种类*/
	card_type, regexpResult := RegexpMatch("<span class=\"item_box_value\" id=\"dCnType\">(?s:(.*?))</span>", body)
	if regexpResult {
		return "数据存在错误: 无法获取卡片种类", ""
	}
	if strings.Contains(card_type, "魔法") || strings.Contains(card_type, "陷阱") {
		ygodata = ygodata + card_type
	} else {
		/*种族*/
		card_race, regexpResult := RegexpMatch("<span id=\"dCnRace\" class=\"item_box_value\">(?s:(.*?))</span>", body)
		if regexpResult {
			return "数据存在错误: 无法获取卡片种族", ""
		}
		ygodata = ygodata + card_race + "族  "
		/*星数*/
		var cardlevel string
		if strings.Contains(card_type, "连接") {
			ygodata = ygodata + "LINK"
			cardlevel, regexpResult = RegexpMatch("LINK(?s:(.*?))</span>", body)
			if regexpResult {
				cardlevel, regexpResult = RegexpMatch("\"item_box_value\">Link(?s:(.*?))</span>", body)
				if regexpResult {
					return "数据存在错误: 无法获取连接数值", ""
				}
			}
		} else {
			if strings.Contains(card_type, "超量") || strings.Contains(card_type, "XYZ") {
				ygodata = ygodata + "阶级："
			} else {
				ygodata = ygodata + "等级："
			}
			cardlevel, regexpResult = RegexpMatch("<b>星数/阶级</b> </span><span class=\"item_box_value\">(?s:(.*?))</span>", body)
			if regexpResult {
				return "数据存在错误: 无法获取等级、阶级", ""
			}
		}
		ygodata = ygodata + cardlevel + "  属性："
		/*属性*/
		card_attr, regexpResult := RegexpMatch("<span class=\"item_box_value\" id=\"attr\">(?s:(.*?))</span>", body)
		if regexpResult {
			return "数据存在错误: 无法获取属性", ""
		}
		ygodata = ygodata + card_attr + "\n"
		/*种类*/
		ygodata = ygodata + card_type + "  atk:"
		/*攻击力*/
		card_atk, regexpResult := RegexpMatch("<b>ATK</b> </span>&nbsp;<span class=\"item_box_value\">(?s:(.*?))</span>", body)
		if regexpResult {
			return "数据存在错误: 无法获取攻击力", ""
		}
		ygodata = ygodata + card_atk
		/*守备力*/
		if !strings.Contains(card_type, "连接") {
			result_def := regexp.MustCompile("<b>DEF</b></span>(?s:(.*?))</span>")
			carddef := result_def.FindAllStringSubmatch(body, -1)
			card_def := strings.Replace(carddef[0][0], "\r\n", "", -1) //去掉空格
			card_def = strings.Replace(card_def, " ", "", -1)          //去掉空格
			card_def2, _ := RegexpMatch("\"item_box_value\">(.*?)</span>", card_def)
			ygodata = ygodata + "  def:" + card_def2
		}
	}
	/*效果*/
	result_depict := regexp.MustCompile("<div class=\"item_box_text\" id=\"cardDepict\">(?s:(.*?))</div>")
	carddepict := result_depict.FindAllStringSubmatch(body, -1)
	card_depict := strings.Replace(carddepict[0][1], " ", "", -1) //去掉空格
	//card_depict = strings.Replace(card_depict, "\r\n", "", -1)         //去掉空格
	ygodata = ygodata + card_depict

	return ygodata, imageBase64
}

//获取字段列表
func getYGolist(cardsnumber string, body string) (ygolist string) {
	//判断是否匹配到东西
	if cardsnumber == "0" {
		ygolist = "未找到卡片，请检查卡名是否正确"
	} else {
		//获取页数最大值
		page_number, _ := RegexpMatch("个卡片, 翻(?s:(.*?))页可看完", string(body))
		pagemax, _ = strconv.Atoi(page_number)
		//输出总张数*/
		ygolist = "找到" + cardsnumber + "张相关卡片,当前显示以下卡名：\n"
		//获取卡名列表
		regsult_name := regexp.MustCompile("height=\"144px\"  alt=\"(?s:(.*?))\" src=\"")
		namelist := regsult_name.FindAllStringSubmatch(body, -1)
		for i, names := range namelist {
			CardsName = append(CardsName, strconv.Itoa(i)+"."+names[1])
		}
		//获取链接列表
		regsult_href := regexp.MustCompile("<a href=\"..(.*?)\" target=\"_blank\">")
		hreflist := regsult_href.FindAllStringSubmatch(body, -1)
		for i, frefs := range hreflist {
			if i > 1 && i%2 == 0 {
				CardsHref = append(CardsHref, frefs[1])
			}
		}
		ygolist = ygolist + strings.Join(CardsName, "\n")
	}
	return ygolist
}
