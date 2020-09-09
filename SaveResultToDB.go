package main

import (
	"bytes"
	"container/list"
	"database/sql"
	"encoding/xml"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

//组名全局变量定义
var groupNameMap = map[int]string {
1: "策划1组",
2: "策划2组",
3: "客户端1组-越界组",
4: "客户端2组-溢出组",
5: "客户端3组-空指针组",
6: "海外版组",
7: "服务器组",
8: "运营1组",
9: "运营2组",
10: "场景1组",
11: "场景2组",
12: "角色1组",
13: "角色2组",
14: "赛车组",
15: "动特1组",
16: "动特2组",
17: "视觉设计组",
18: "测试组",
19: "海外测试组",
20: "端游",
21: "X7运营组",
}

//Group 数据 BEGIN
type GroupInfo struct {
	m_groupID int
	m_groupName string
	m_groupPlayerNum int
	m_joinRate int
	m_score float32
}

type OrderedGroupInfoList []GroupInfo

func (p OrderedGroupInfoList) Len() int {
	return len(p)
}

func (p OrderedGroupInfoList) Less(i, j int) bool {
	if p[i].m_joinRate != p[j].m_joinRate {
		return p[i].m_joinRate > p[j].m_joinRate
	}

	return p[i].m_score > p[j].m_score
}

func (p OrderedGroupInfoList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p *OrderedGroupInfoList) Sort() {
	sort.Sort(p)
}
//Group 数据 END

//Player 数据 BEGIN
type PlayerInfo struct {
	m_playerUID string
	m_playerName string
	m_playerScore int
	m_playerJoinNum int
	m_playerGroupID int
	m_playerTotalGameTime int
}

type OrderedPlayerInfoList []PlayerInfo

func (p OrderedPlayerInfoList) Len() int {
	return len(p)
}
func (p OrderedPlayerInfoList) Less(i, j int) bool {
	if p[i].m_playerScore != p[j].m_playerScore {
		return p[i].m_playerScore > p[j].m_playerScore
	}

	return p[i].m_playerTotalGameTime > p[j].m_playerTotalGameTime
}

func (p OrderedPlayerInfoList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p *OrderedPlayerInfoList) Sort() {
	sort.Sort(p)
}
//Player 数据 END

func main() {
	if (len(os.Args) != 2) {
		fmt.Printf("Usage:\n%s 20200101.xml\n", os.Args[0])
		return;
	}

	fileList := list.New()
	for idx, arg := range os.Args {
		//fmt.Printf("main arg(%d), value(%s)\n", idx, arg)
		if (idx >= 1) {
			fileList.PushBack(string(arg))
		}
	}
	fmt.Printf("file list size|%d|\n", fileList.Len())
	fmt.Printf("file name|%s|\n", os.Args[1])
	fileName := os.Args[1]
	tmpList := strings.Split(fileName, ".");
	match_id := tmpList[0]
	fmt.Printf("match id|%s|\n", match_id)

	//读取DB, 校验这个match id是否已经处理过了
	db, err := sql.Open("sqlite3", "./db.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("select match_id from polls_matchid")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var db_match_id string
		err = rows.Scan(&db_match_id)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Printf("found match id in DB: %s\n", db_match_id)
		if db_match_id == match_id {
			fmt.Printf("match id have been insert in DB, please check: %s !!!\n", db_match_id)
			//WEIFIXME, 测试的时候把这个关闭
			return
		}
	}

	//插入MatchID表
	txInsertMatchID, errDB := db.Begin()
	if errDB != nil {
		log.Fatal(errDB)
	}
	stmtInsertMatchID, errDB := txInsertMatchID.Prepare("insert into polls_matchid(match_id) values(?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmtInsertMatchID.Close()
	_, err = stmtInsertMatchID.Exec(match_id)
	if err != nil {
		log.Fatal(err)
	}
	txInsertMatchID.Commit()


	//--------------------------解析xml--------------------------------------
	groupInfoMap := make(map[int]GroupInfo)
	playerInfoMap := make(map[string]PlayerInfo)

	totalGameCnt := 0
	roundNumThisGame := 0

	for fileItem := fileList.Front(); fileItem != nil; fileItem = fileItem.Next() {
		fileName := fileItem.Value
		content, err := ioutil.ReadFile(fileName.(string))
		if (err != nil) {
			fmt.Printf("open xml file fail|%v|\n", err)
			return
		}

		totalGameCnt++

		var t xml.Token
		//遍历所有xml的元素
		decoder := xml.NewDecoder(bytes.NewBuffer(content))
		for t, err = decoder.Token(); err == nil; t, err = decoder.Token() {
			switch token := t.(type) {
			case xml.StartElement:
				name := token.Name.Local
				if name == "InnerMatch" {
					for _, attr := range token.Attr {
						attrName := attr.Name.Local
						attrValue := attr.Value

						if attrName == "CurRound" {
							roundNumThisGame, err = strconv.Atoi(attrValue)
							if err != nil {
								fmt.Printf("conv CurRound fail: %s", err)
								return
							}
							fmt.Printf("found curround: %d\n", roundNumThisGame)
						}
					}
				} else if name == "Group" {
					//fmt.Printf("Token name: %s\n", name)

					//解析出group结构体
					var groupItem GroupInfo
					for _, attr := range token.Attr {
						attrName := attr.Name.Local
						attrValue := attr.Value
						fmt.Printf("group \n\n%s = %s\n", attrName, attrValue)

						if attrName == "GroupID" {
							groupItem.m_groupID, err = strconv.Atoi(attrValue)
							if err != nil {
								fmt.Printf("conv groupID fail: %s", err)
								return
							}
							groupItem.m_groupName = groupNameMap[groupItem.m_groupID]
						} else if attrName == "GroupPlayerNum" {
							groupItem.m_groupPlayerNum, err = strconv.Atoi(attrValue)
							if err != nil {
								fmt.Printf("conv m_groupPlayerNum fail: %s", err)
								return
							}
						} else if attrName == "JoinRate" {
							groupItem.m_joinRate, err = strconv.Atoi(attrValue)
							if err != nil {
								fmt.Printf("conv JoinRate fail: %s", err)
								return
							}
						} else if attrName == "Score" {
							tmpScore, err := strconv.Atoi(attrValue)
							groupItem.m_score = float32(tmpScore)
							if err != nil {
								fmt.Printf("conv Score fail: %s", err)
								return
							}
						}
					} // end for

					//分数要算成个人平均分
					if groupItem.m_groupPlayerNum != 0 {
						groupItem.m_score = groupItem.m_score / float32(groupItem.m_groupPlayerNum)
					} else {
						groupItem.m_score = 0.
					}

					//确认元素是否已经存在
					var /*thisGroupInfo*/ _, exist = groupInfoMap[groupItem.m_groupID]
					if exist {
						/*fmt.Printf("old joinrate(%d), old score(%d)\n", thisGroupInfo.m_joinRate, thisGroupInfo.m_score)
						thisGroupInfo.m_score += groupItem.m_score
						thisGroupInfo.m_joinRate += groupItem.m_joinRate
						fmt.Printf("new joinrate(%d), new score(%d)\n", thisGroupInfo.m_joinRate, thisGroupInfo.m_score)

						//必须重新加入map
						//https://stackoverflow.com/questions/17438253/accessing-struct-fields-inside-a-map-value-without-copying
						groupInfoMap[groupItem.m_groupID] = thisGroupInfo */
					} else {
						fmt.Printf("group first add\n")
						groupInfoMap[groupItem.m_groupID] = groupItem
					}

					fmt.Printf("group(%d) joinRate(%d) score(%d)\n", groupItem.m_groupID, groupInfoMap[groupItem.m_groupID].m_joinRate, float32(groupInfoMap[groupItem.m_groupID].m_score))
				} else if name == "Player" {
					//解析出player结构体
					var playerItem PlayerInfo
					fmt.Printf("player token size: %d\n", len(token.Attr))

					for _, attr := range token.Attr {
						attrName := attr.Name.Local
						attrValue := attr.Value
						//fmt.Printf("player \n\n%s = %s\n", attrName, attrValue)

						if attrName == "UID" {
							playerItem.m_playerUID = attrValue
							if err != nil {
								fmt.Printf("conv UID fail: %s", err)
								return
							}
						} else if attrName == "GroupID" {
							playerItem.m_playerGroupID, err = strconv.Atoi(attrValue)
							if err != nil {
								fmt.Printf("conv GroupID fail: %s", err)
								return
							}
							fmt.Printf("found group id: %d\n", playerItem.m_playerGroupID)
						} else if attrName == "Name" {
							playerItem.m_playerName = attrValue
							fmt.Printf("find player name: %s\n", playerItem.m_playerName)
						} else if attrName == "JoinNum" {
							playerItem.m_playerJoinNum, err = strconv.Atoi(attrValue)
							if err != nil {
								fmt.Printf("conv JoinNum fail: %s", err)
								return
							}
						} else if attrName == "Score" {
							playerItem.m_playerScore, err = strconv.Atoi(attrValue)
							if err != nil {
								fmt.Printf("conv Score fail: %s", err)
								return
							}
							fmt.Printf("find player score: %d\n", playerItem.m_playerScore)
						} else if attrName == "JoinNum" {
							playerItem.m_playerJoinNum, err = strconv.Atoi(attrValue)
							if err != nil {
								fmt.Printf("conv JoinNum fail: %s", err)
								return
							}
						} else if attrName == "TotalGameTime" {
							playerItem.m_playerTotalGameTime, err = strconv.Atoi(attrValue)
							if err != nil {
								fmt.Printf("conv TotalGameTime fail: %s", err)
								return
							}
						}
						//fmt.Printf("from config, player name: %s, score: %d\n", playerItem.m_playerName, playerItem.m_playerScore)
					} // end for

					//确认元素是否已经存在
					var /*thisGroupInfo*/ _, exist = playerInfoMap[playerItem.m_playerUID]
					if exist {
						/*fmt.Printf("old joinrate(%d), old score(%d)\n", thisGroupInfo.m_joinRate, thisGroupInfo.m_score)
						thisGroupInfo.m_score += groupItem.m_score
						thisGroupInfo.m_joinRate += groupItem.m_joinRate
						fmt.Printf("new joinrate(%d), new score(%d)\n", thisGroupInfo.m_joinRate, thisGroupInfo.m_score)

						//必须重新加入map
						//https://stackoverflow.com/questions/17438253/accessing-struct-fields-inside-a-map-value-without-copying
						groupInfoMap[groupItem.m_groupID] = thisGroupInfo */
					} else {
						fmt.Printf("player first add, player name: %s, score: %d\n", playerItem.m_playerName, playerItem.m_playerScore)
						playerInfoMap[playerItem.m_playerUID] = playerItem
					}
				}
			// 处理元素结束（标签）
			case xml.EndElement:
				//fmt.Printf("Token of '%s' end\n\n", token.Name.Local)
			// 处理字符数据（这里就是元素的文本）
			case xml.CharData:
				//content := string([]byte(token))
				//fmt.Printf("This is the content: %v\n", content)
			default:
				// ...

			}
		}// end for xml
	}// end for file list

	fmt.Printf("-----------------totalGameCount(%d)\n", totalGameCnt)

	//1.首先 处理group表
	var sortResult OrderedGroupInfoList
	for _,value := range groupInfoMap {
		sortResult = append(sortResult, value)
	}
	sortResult.Sort()

	//写DB
	txGroupPlayer, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmtGroupPlayer, err := txGroupPlayer.Prepare("insert into polls_group(match_id, group_id, group_name, group_rank, group_join_rate, group_avg_score) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmtGroupPlayer.Close()
	//插入数据
	groupRank := 0
	for i := 0; i < len(sortResult); i++ {
		if sortResult[i].m_groupID == 0 {
			continue;
		}

		groupRank++
		groupID := sortResult[i].m_groupID
		groupName := sortResult[i].m_groupName
		groupJoinRate := sortResult[i].m_joinRate
		groupAvgScore := sortResult[i].m_score
		fmt.Printf("group id(%d), group name(%s), total join rate(%d), avg score(%f)\n", groupID, groupName, groupJoinRate, groupAvgScore)

		_, err = stmtGroupPlayer.Exec(match_id, groupID, groupName, groupRank, groupJoinRate, groupAvgScore)
		if err != nil {
			log.Fatal(err)
		}
	}
	//tx.Commit()

	//---------------------2.然后处理Player表---------------------------------
	var sortPlayerResult OrderedPlayerInfoList
	for _,value := range playerInfoMap {
		//fmt.Printf("before add to sort player, name: %s, score: %d\n", value.m_playerName, value.m_playerScore)
		sortPlayerResult = append(sortPlayerResult, value)
	}
	sortPlayerResult.Sort()
	fmt.Printf("player list size: %d\n", len(sortPlayerResult))

	//写DB
	stmtPlayer, err2 := txGroupPlayer.Prepare("insert into polls_player(match_id, player_rank, player_uid, player_name, player_total_score, player_join_num, player_join_rate, player_group_id, player_total_game_time) values(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err2 != nil {
		log.Fatal(err2)
	}
	defer stmtGroupPlayer.Close()

	//插入数据
	playerRank := 0
	for i := 0; i < len(sortPlayerResult); i++ {
		playerRank++
		playerGroupID := sortPlayerResult[i].m_playerGroupID
		playerUID := sortPlayerResult[i].m_playerUID
		playerName := sortPlayerResult[i].m_playerName
		playerTotalScore := sortPlayerResult[i].m_playerScore
		playerJoinNum := sortPlayerResult[i].m_playerJoinNum
		playerJoinRate := float32(sortPlayerResult[i].m_playerJoinNum) / float32(roundNumThisGame) * 10000
		playerTotalGameTime := sortPlayerResult[i].m_playerTotalGameTime
		fmt.Printf("UID: %s, name: %s, total score: %d, join num: %d, join rate: %d, group id: %d, game time: %d\n", playerUID, sortPlayerResult[i].m_playerName, sortPlayerResult[i].m_playerScore, playerJoinNum, int32(playerJoinRate), playerGroupID, playerTotalGameTime)

		_, err = stmtPlayer.Exec(match_id, playerRank, playerUID, playerName, playerTotalScore, playerJoinNum, int32(playerJoinRate), playerGroupID, playerTotalGameTime)
		if err != nil {
			log.Fatal(err)
		}
	}
	txGroupPlayer.Commit()

	//检查结果

	/*stmtGroupPlayer, err = db.Prepare("select group_name from polls_group where match_id = ? and group_id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmtGroupPlayer.Close()
	var group_name string
	err = stmtGroupPlayer.QueryRow(match_id, 1).Scan(&group_name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("group name: " + group_name)*/

	fmt.Println("save to DB succ\n")
}
