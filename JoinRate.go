package main

import (
	"bytes"
	"container/list"
	"encoding/xml"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"time"
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

type GroupInfo struct {
	m_groupID int
	m_groupPlayerNum int
	m_joinRate int
	m_score int
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

func main() {
	if (len(os.Args) <= 1) {
		fmt.Printf("Usage:\n%s file1.xml file2.xml file3.xml ...\n", os.Args[0])
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

	groupInfoMap := make(map[int]GroupInfo)

	totalGameCnt := 0
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
				//只处理group
				if name != "Group" {
					continue
				}

				fmt.Printf("Token name: %s\n", name)

				//解析出结构体
				var groupItem GroupInfo
				for _, attr := range token.Attr {
					attrName := attr.Name.Local
					attrValue := attr.Value
					fmt.Printf("\n\n%s = %s\n", attrName, attrValue)

					if attrName == "GroupID" {
						groupItem.m_groupID, err = strconv.Atoi(attrValue)
						if err != nil {
							fmt.Printf("conv groupID fail: %s", err)
						}
					} else if attrName == "GroupPlayerNum" {
						groupItem.m_groupPlayerNum, err = strconv.Atoi(attrValue)
						if err != nil {
							fmt.Printf("conv m_groupPlayerNum fail: %s", err)
						}
					} else if attrName == "JoinRate" {
						groupItem.m_joinRate, err = strconv.Atoi(attrValue)
						if err != nil {
							fmt.Printf("conv JoinRate fail: %s", err)
						}
					} else if attrName == "Score" {
						groupItem.m_score, err = strconv.Atoi(attrValue)
						if err != nil {
							fmt.Printf("conv Score fail: %s", err)
						}
					}
				} // end for

				//分数要算成个人平均分
				if groupItem.m_groupPlayerNum != 0 {
					groupItem.m_score = groupItem.m_score / groupItem.m_groupPlayerNum
				} else {
					groupItem.m_score = 0
				}

				//确认元素是否已经存在
				var thisGroupInfo, exist = groupInfoMap[groupItem.m_groupID]
				if exist {
					fmt.Printf("old joinrate(%d), old score(%d)\n", thisGroupInfo.m_joinRate, thisGroupInfo.m_score)
					thisGroupInfo.m_score += groupItem.m_score
					thisGroupInfo.m_joinRate += groupItem.m_joinRate
					fmt.Printf("new joinrate(%d), new score(%d)\n", thisGroupInfo.m_joinRate, thisGroupInfo.m_score)
					//https://stackoverflow.com/questions/17438253/accessing-struct-fields-inside-a-map-value-without-copying
					groupInfoMap[groupItem.m_groupID] = thisGroupInfo
				} else {
					fmt.Printf("first add\n")
					groupInfoMap[groupItem.m_groupID] = groupItem
				}

				fmt.Printf("group(%d) joinRate(%d) score(%d)\n", groupItem.m_groupID, groupInfoMap[groupItem.m_groupID].m_joinRate, groupInfoMap[groupItem.m_groupID].m_score)

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
	var sortResult OrderedGroupInfoList
	for _,value := range groupInfoMap {
		sortResult = append(sortResult, value)
	}
	sortResult.Sort()

	//写到xls
	xlsx := excelize.NewFile()
	//设置标题
	xlsx.SetCellValue("Sheet1", "A1", "组名")
	xlsx.SetCellValue("Sheet1", "B1", "参与率")
	xlsx.SetCellValue("Sheet1", "C1", "平均分")

	fmt.Printf("total game count(%d)\n", totalGameCnt)
	rowNum := 1
	//for key,value := range groupInfoMap {
	for i := 0; i < len(sortResult); i++ {
		groupID := sortResult[i].m_groupID
		groupName := groupNameMap[groupID]
		fmt.Printf("total join rate(%d)\n", sortResult[i].m_joinRate)
		avgJoinRate := float32(sortResult[i].m_joinRate) / float32(totalGameCnt) / 100.
		avgScore := float32(sortResult[i].m_score) / float32(totalGameCnt) // / float32(sortResult[i].m_groupPlayerNum)

		fmt.Printf("group id(%d), name(%s), join rate(%f), score(%f)\n", groupID, groupName, avgJoinRate, avgScore)

		rowNum++
		err := xlsx.SetCellValue("Sheet1", fmt.Sprintf("A%d", rowNum), groupName)
		if err != nil {
			fmt.Printf("save groupName fail,%v\n", err)
			return
		}

		xlsx.SetCellValue("Sheet1", fmt.Sprintf("B%d", rowNum), strconv.FormatFloat(float64(avgJoinRate), 'f',-2,32))
		if err != nil {
			fmt.Printf("save avgJoinRate fail,%v\n", err)
			return
		}

		xlsx.SetCellValue("Sheet1", fmt.Sprintf("C%d", rowNum), strconv.FormatFloat(float64(avgScore), 'f',-2,32))
		if err != nil {
			fmt.Printf("save avgScore fail,%v\n", err)
			return
		}
	}

	err := xlsx.SaveAs(fmt.Sprintf("./month_%d.xlsx", time.Now().Unix()))
	if err != nil {
		fmt.Println("save final xls fail: %v\n", err)
	}
	fmt.Println("save to xml succ\n")
}