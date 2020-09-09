package main

import (
	//"bytes"
	"container/list"
	"time"

	//"encoding/xml"
	"fmt"
	"github.com/beevik/etree"
	//"io/ioutil"
	"os"
	//"sort"
	"strconv"
	//"time"
)

type GroupInfo struct {
	m_groupID int
	m_groupPlayerNum int
	m_joinRate int
	m_score int
}

const kOpeTypeModifyMap int = 1
const kOpeTypeAfk int = 2
const kOpeTypeSetBeginTime int = 3


const kSpeedRaceGamePlay int = 4
const kPropRaceGamePlay int = 9
const kSpeedRaceRoundNum int = 4

func main() {
	if (len(os.Args) < 2) {
		fmt.Printf("Usage:\n %s OperateType arg1 arg2...\n", os.Args[0])
		os.Exit(1)
	}

	opType, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("first arg must be int, %v\n", err)
		os.Exit(1)
	}

	//读取配置
	doc := etree.NewDocument()
	err = doc.ReadFromFile("InnerMatch.xml")
	if err != nil {
		fmt.Printf("read Cfg file fail, %v\n", err)
	}
	innerGameNode := doc.SelectElement("InnerGame")

	//修改地图
	if opType == kOpeTypeModifyMap {
		mapIDList := list.New()
		for idx, arg := range os.Args {
			if (idx < 2) {
				continue
			}

			//把地图存下来
			var mapID int
			mapID, err = strconv.Atoi(arg)
			if err != nil {
				fmt.Println("parse map id fail, idx(%d), %v\n", idx + 1, err)
				os.Exit(1)
			}
			mapIDList.PushBack(mapID)
		}

		mapIDListSize := mapIDList.Len()
		fmt.Printf("map id num(%d)\n", mapIDListSize)

		//开始修改RoundCfg
		roundCfgRootNode := innerGameNode.SelectElement("RoundCfg")
		for _, roundItem := range roundCfgRootNode.SelectElements("Round") {
			roundCfgRootNode.RemoveChild(roundItem)
		}

		mapIdx := 0
		for mapID := mapIDList.Front(); mapID != nil; mapID = mapID.Next() {
			newRound := etree.NewElement("Round")
			newRound.CreateAttr("MapID", fmt.Sprintf("%d", mapID.Value))
			var raceGamePlay int
			if mapIdx < kSpeedRaceRoundNum {
				raceGamePlay = kSpeedRaceGamePlay
			} else {
				raceGamePlay = kPropRaceGamePlay
			}
			newRound.CreateAttr("RaceGameType", fmt.Sprintf("%d", raceGamePlay))

			roundCfgRootNode.AddChild(newRound)
			mapIdx++
		}

		fmt.Printf("Confiure Map Succ\n")

	} else if opType == kOpeTypeAfk {
		afkPlayerNameList := list.New()
		for idx, arg := range os.Args {
			if (idx < 2) {
				continue
			}

			//请假名单存下来
			afkPlayerNameList.PushBack(arg)
		}

		playerNum := afkPlayerNameList.Len()
		fmt.Printf("afk player num(%d)\n", playerNum)

		//请假
		playerCfgRootNode := innerGameNode.SelectElement("PlayerCfg")
		for _, playerItem := range playerCfgRootNode.SelectElements("Player") {
			playerName := playerItem.SelectAttrValue("Name", "NoBody")

			//重置
			playerItem.CreateAttr("IsAfk", "false")

			//看是否在请假列表中
			for afkPlayerName := afkPlayerNameList.Front(); afkPlayerName != nil; afkPlayerName = afkPlayerName.Next() {
				if playerName == afkPlayerName.Value {
					playerItem.CreateAttr("IsAfk", "true")
					break
				}
			}
		}

	} else  if opType == kOpeTypeSetBeginTime {
		if (len(os.Args) < 4) {
			fmt.Printf("BeginTime Not given\n")
			os.Exit(1)
		}

		strBeginTime := os.Args[2] + " " + os.Args[3]
		fmt.Printf("set begin time(%s)\n", strBeginTime)

		//格式安全校验
		_, err := time.Parse("2006-01-02 15:04:05", strBeginTime)
		if err != nil {
			fmt.Printf("begin time format error\n")
			os.Exit(1)
		}

		//设置开始时间
		innerGameNode.CreateAttr("BeginTime", strBeginTime)

		//match id递增
		curMatchID, err := strconv.Atoi(innerGameNode.SelectAttrValue("MatchID", "999"))
		if err != nil {
			fmt.Printf("old match id format err\n")
			os.Exit(1)
		}
		newMatchID := curMatchID + 1
		innerGameNode.CreateAttr("MatchID", fmt.Sprintf("%d", newMatchID))

		//设置所有人为不请假
		playerCfgRootNode := innerGameNode.SelectElement("PlayerCfg")
		for _, playerItem := range playerCfgRootNode.SelectElements("Player") {
			playerItem.CreateAttr("IsAfk", "false")
		}

	} else {
		fmt.Printf("不支持的操作类型(%d)\n", opType)
		os.Exit(1)
	}

	doc.Indent(2)
	doc.WriteToFile(fmt.Sprintf("InnerMatch.xml")) //, time.Now().Unix()))
	fmt.Println("save to xml succ\n")
}