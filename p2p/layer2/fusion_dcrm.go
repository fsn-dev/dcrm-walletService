package layer2

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"github.com/fsn-dev/dcrm-walletService/crypto"
	"github.com/fsn-dev/dcrm-walletService/internal/common"
	"github.com/fsn-dev/dcrm-walletService/p2p/discover"
)

const (
	timeout_round = 60//second
)

func init() {
	//RegisterRecvCallback(Round_5)
	SdkProtocol_registerBroadcastInGroupCallback(Round_5)
	SdkProtocol_registerSendToGroupCallback(reqAddr_start)
	SdkProtocol_registerSendToGroupReturnCallback(reqAddr_finish)
	RegisterCallback(Round_5)
}

func GetAddr(kid string) (string, [6]int)  {
	key_value_lock.Lock()
	defer key_value_lock.Unlock()
	if key_value[kid] != nil {
		return key_value[kid].success, key_value[kid].round
	}
	return "false", key_value[kid].round
}

func ReqAddr(gid string) string {
	fmt.Printf("==== ReqAddr() ====, gid = %v\n", gid)
	//send reqaddr
	key := crypto.Keccak256Hash([]byte(fmt.Sprintf("%v", time.Now().UnixNano() / 1e6))).Hex()
	keyhex := fmt.Sprintf("%v", key)
	keyid := keyhex[2:]
	fmt.Printf("==== ReqAddr() ====, key: %v\n", keyid)
	msg := fmt.Sprintf("%v:0:%v-%v", keyid, gid, msg_1300)
	go SdkProtocol_SendToGroupAllNodes(gid, msg)//start
	return keyid// hash
}

func reqAddr_start(msg interface{}, nodeid string) <-chan string {
	fmt.Printf("==== reqAddr_start() ====, nodeid: %v, msg = %v\n", nodeid, msg.(string))
	slice1 := strings.Split(msg.(string), "-")
	if len(slice1) < 2 {
		fmt.Printf("==== reqAddr_start() ====, nodeid: %v, msg = %v format '-' err\n", nodeid, msg.(string))
		ret := make(chan string)
		return ret
	}
	slice2 := strings.Split(slice1[0], ":")
	if len(slice2) < 3 {
		fmt.Printf("==== reqAddr_start() ====, nodeid: %v, msg = %v format ':' err\n", nodeid, msg.(string))
		ret := make(chan string)
		return ret
	}
	onceLock.Lock()
	if once == false {
		groupid, _ := HexID(slice2[2])
		nodes := getSDKGroupNodes(groupid)
		var nid []string = make([]string, 0)
		for _, n := range nodes {
			nid = append(nid, n.ID.String())
		}
		sort.Sort(sort.StringSlice(nid))
		for i, id := range nid {
			nodeid_value[id] = 1 << i
			fmt.Printf("==== reqAddr_start() ====, gid: %v, nodeid_value[%v]: %v\n", slice2[2], id, nodeid_value[id])
		}
		once = true
		SelfID = fmt.Sprintf("%v", discover.GetLocalID())
	}
	onceLock.Unlock()
	key_value_lock.Lock()
	if key_value[slice2[0]] == nil {
		key_value[slice2[0]] = &key5{finish: make(chan string), timeAt: time.Now().Unix()}
	} else {
		key_value[slice2[0]].timeAt = time.Now().Unix()
	}
	fmt.Printf("%v ==== reqAddr_start() ====, key: %v, timeAt: %v\n", common.CurrentTime(), slice2[0], key_value[slice2[0]].timeAt)
	key_value_lock.Unlock()
	if key_value[slice2[0]].future[1] != nil {
		for i, s := range key_value[slice2[0]].future[1].msg {
			go Round_15(s, i, false)
			defer delete(key_value[slice2[0]].future[1].msg, i)
		}
	}
	//time.Sleep(time.Duration(2) * time.Second)
	msg2 := fmt.Sprintf("%v:1:%v-%v", slice2[0], slice2[2], slice1[1])
	go SdkProtocol_broadcastInGroupOthers(slice2[2], msg2)
	fmt.Printf("==== reqAddr_start() ====, key: %v, finish wait\n", slice2[0])
	ret := make(chan string)
	//fstring := <-key_value[slice2[0]].finish
	//fmt.Printf("==== ReqAddr() ====, key: %v, finish: %v\n", slice2[0], fstring)
	//ret <- fstring
	return ret
}

func Round_5(msg interface{}, nodeid string) {
	go Round_15(msg, nodeid, true)
}

func Round_15(msg interface{}, nodeid string, recursive bool) {
	recursive = true// TODO test delete
	fmt.Printf("==== Round_15() ====, msg = %v\n", msg.(string))
	slice1 := strings.Split(msg.(string), "-")
	slice2 := strings.Split(slice1[0], ":")
	onceLock.Lock()
	if once == false {
		groupid, _ := HexID(slice2[2])
		nodes := getSDKGroupNodes(groupid)
		var nid []string = make([]string, 0)
		for _, n := range nodes {
			nid = append(nid, n.ID.String())
		}
		sort.Sort(sort.StringSlice(nid))
		for i, id := range nid {
			nodeid_value[id] = 1 << i
			fmt.Printf("==== Round_15() ====, gid: %v, nodeid_value[%v]: %v\n", slice2[2], id, nodeid_value[id])
		}
		once = true
		SelfID = fmt.Sprintf("%v", discover.GetLocalID())
	}
	onceLock.Unlock()
	key_value_lock.Lock()
	if key_value[slice2[0]] == nil {
		key_value[slice2[0]] = &key5{finish: make(chan string)}
	}
	key_value_lock.Unlock()
	fmt.Printf("==== Round_15() ====, msg = %v, recv round: %v\n", msg.(string), slice2[1])
	switch slice2[1] {
	case "1":
		ret := store_msg(1, slice2[0], msg, nodeid)
		if ret == true {
			key_value_lock.Lock()
			at := key_value[slice2[0]].timeAt
			key_value_lock.Unlock()
			if time.Now().Unix() - at > timeout_round {
				fmt.Printf("==== Round_15(1) ====, keyid: %v, at: %v, now: %v, timeout round\n", slice2[0], at, time.Now().Unix())
				break
			}
			key_value_lock.Lock()
			key_value[slice2[0]].timeAt = time.Now().Unix()
			fmt.Printf("==== Round_15(1) ====, key: %v, timeAt: %v\n", slice2[0], key_value[slice2[0]].timeAt)
			key_value_lock.Unlock()
			if recursive == true && key_value[slice2[0]].future[2] != nil {
				for i, s := range key_value[slice2[0]].future[2].msg {
					go Round_15(s, i, false)
					defer delete(key_value[slice2[0]].future[2].msg, i)
				}
			}
			//time.Sleep(time.Duration(1) * time.Second)
			go Round_2_SHARE1(msg.(string))
		}
		break
	case "2":
		ret := store_msg(2, slice2[0], msg, nodeid)
		if ret == true {
			key_value_lock.Lock()
			at := key_value[slice2[0]].timeAt
			key_value_lock.Unlock()
			if time.Now().Unix() - at > timeout_round {
				fmt.Printf("==== Round_15(2) ====, keyid: %v, at: %v, now: %v, timeout round\n", slice2[0], at, time.Now().Unix())
				break
			}
			key_value_lock.Lock()
			key_value[slice2[0]].timeAt = time.Now().Unix()
			fmt.Printf("==== Round_15(2) ====, key: %v, timeAt: %v\n", slice2[0], key_value[slice2[0]].timeAt)
			key_value_lock.Unlock()
			if recursive == true && key_value[slice2[0]].future[3] != nil {
				for i, s := range key_value[slice2[0]].future[3].msg {
					go Round_15(s, i, false)
					defer delete(key_value[slice2[0]].future[3].msg, i)
				}
			}
			//time.Sleep(time.Duration(1) * time.Second)
			msg2 := fmt.Sprintf("%v:3:%v-%v", slice2[0], slice2[2], slice1[1])
			go SdkProtocol_broadcastInGroupOthers(slice2[2], msg2)
		}
		break
	case "3":
		ret := store_msg(3, slice2[0], msg, nodeid)
		if ret == true {
			key_value_lock.Lock()
			at := key_value[slice2[0]].timeAt
			key_value_lock.Unlock()
			if time.Now().Unix() - at > timeout_round {
				fmt.Printf("==== Round_15(3) ====, keyid: %v, at: %v, now: %v, timeout round\n", slice2[0], at, time.Now().Unix())
				break
			}
			key_value_lock.Lock()
			key_value[slice2[0]].timeAt = time.Now().Unix()
			fmt.Printf("==== Round_15(3) ====, key: %v, timeAt: %v\n", slice2[0], key_value[slice2[0]].timeAt)
			key_value_lock.Unlock()
			if recursive == true && key_value[slice2[0]].future[4] != nil {
				for i, s := range key_value[slice2[0]].future[4].msg {
					go Round_15(s, i, false)
					defer delete(key_value[slice2[0]].future[4].msg, i)
				}
			}
			//time.Sleep(time.Duration(1) * time.Second)
			msg2 := fmt.Sprintf("%v:4:%v-%v", slice2[0], slice2[2], slice1[1])
			go SdkProtocol_broadcastInGroupOthers(slice2[2], msg2)
		}
		break
	case "4":
		ret := store_msg(4, slice2[0], msg, nodeid)
		if ret == true {
			key_value_lock.Lock()
			at := key_value[slice2[0]].timeAt
			key_value_lock.Unlock()
			if time.Now().Unix() - at > timeout_round {
				fmt.Printf("==== Round_15(4) ====, keyid: %v, at: %v, now: %v, timeout round\n", slice2[0], at, time.Now().Unix())
				break
			}
			key_value_lock.Lock()
			key_value[slice2[0]].timeAt = time.Now().Unix()
			fmt.Printf("==== Round_15(4) ====, key: %v, timeAt: %v\n", slice2[0], key_value[slice2[0]].timeAt)
			key_value_lock.Unlock()
			if recursive == true && key_value[slice2[0]].future[5] != nil {
				for i, s := range key_value[slice2[0]].future[5].msg {
					go Round_15(s, i, false)
					defer delete(key_value[slice2[0]].future[5].msg, i)
				}
			}
			//time.Sleep(time.Duration(1) * time.Second)
			msg2 := fmt.Sprintf("%v:5:%v-%v", slice2[0], slice2[2], slice1[1])
			go SdkProtocol_broadcastInGroupOthers(slice2[2], msg2)
		}
		break
	case "5":
		ret := store_msg(5, slice2[0], msg, nodeid)
		if ret == true {
			key_value_lock.Lock()
			at := key_value[slice2[0]].timeAt
			key_value_lock.Unlock()
			if time.Now().Unix() - at > timeout_round {
				fmt.Printf("==== Round_15(5) ====, keyid: %v, at: %v, now: %v, timeout round\n", slice2[0], at, time.Now().Unix())
				break
			}
			key_value_lock.Lock()
			key_value[slice2[0]].timeAt = time.Now().Unix()
			fmt.Printf("==== Round_15(5) ====, key: %v, timeAt: %v\n", slice2[0], key_value[slice2[0]].timeAt)
			//msg2 := fmt.Sprintf("%v:finish:%v-%v", slice2[0], slice2[2], slice1[1])
			//key_value[slice2[0]].finish <- msg2
			key_value[slice2[0]].success = "true"
			key_value_lock.Unlock()
		}
		break
	}
}

func Round_2_SHARE1(msg string) {
	fmt.Printf("==== Round_2_SHARE1() ====, msg = %v\n", msg)
	slice1 := strings.Split(msg, "-")
	slice2 := strings.Split(slice1[0], ":")
	if slice2[1] != "1" {
		fmt.Printf("==== Round_2_SHARE1() ====, slice2[1]: %v != '1', msg = %v\n", slice2[1], msg)
		return
	}
	msg2 := fmt.Sprintf("%v:2:%v-%v", slice2[0], slice2[2], slice1[1])
	groupid, _ := HexID(slice2[2])
	nodes := getSDKGroupNodes(groupid)
	for _, node := range nodes {
		fmt.Printf("==== Round_2_SHARE1() ====, nodeid: %v, SelfID: %v, key: %v\n", node.ID.String(), SelfID, slice2[0])
		if node.ID.String() == SelfID {
			continue
		}
		enode := fmt.Sprintf("enode://%v@%v:%v", node.ID, node.IP, node.UDP)
		fmt.Printf("==== Round_2_SHARE1() ====, call SendMsgToPeer enode: %v, msg: %v\n", enode, msg2)
		go SendMsgToPeer(enode, msg2)// 2
	}
}

func store_msg(num int, key string, msg interface{}, nodeid string) bool {
	fmt.Printf("\n==== store_msg() ====, key: %v, num: %v, nodeid_value[%v]: %v\n", key, num, nodeid, nodeid_value[nodeid])
	ret := false
	key_value_lock.Lock()
	if num == 1 {
		if key_value[key].timeAt == 0 {
			if key_value[key].future[num] == nil {
				key_value[key].future[num] = &msgfuture{msg: make(map[string]string)}
			}
			key_value[key].future[num].msg[nodeid] = msg.(string)
			fmt.Printf("==== store_msg() ====, key: %v, future[%v].msg[%v]: %v\n", key, num, nodeid, key_value[key].future[num].msg[nodeid])
			key_value_lock.Unlock()
			return false
		}
	} else {
		if key_value[key].round[num - 1] != 7 {
			if key_value[key].future[num] == nil {
				key_value[key].future[num] = &msgfuture{msg: make(map[string]string)}
			}
			key_value[key].future[num].msg[nodeid] = msg.(string)
			fmt.Printf("==== store_msg() ====, key: %v, future[%v].msg[%v]: %v\n", key, num, nodeid, key_value[key].future[num].msg[nodeid])
			key_value_lock.Unlock()
			return false
		}
	}
	if (key_value[key].round[num] & nodeid_value[nodeid]) == 0 {
		key_value[key].round[num] |= nodeid_value[nodeid]
	}
	fmt.Printf("==== store_msg() ====, key: %v, num: %v, nodeid: %v, key_value[key].rount[%v]: %v, nodeid_value[%v]: %v\n", key, num, nodeid, num, key_value[key].round[num], SelfID, nodeid_value[SelfID])
	key_value[key].round[num] |= nodeid_value[SelfID]
	if key_value[key].round[num] == 7 {
		ret = true
	}
	fmt.Printf("==== store_msg() ====, key: %v, num: %v, nodeid: %v, key_value[key].rount[%v]: %v, ret: %v\n", key, num, nodeid, num, key_value[key].round[num], ret)
	key_value_lock.Unlock()
	return ret
}

func reqAddr_finish(msg interface{}, enode string) {
	fmt.Printf("==== reqAddr_finish() ====, msg = %v\n", msg.(string))
	slice1 := strings.Split(msg.(string), "-")
	slice2 := strings.Split(slice1[0], ":")
	key_value_lock.Lock()
	defer key_value_lock.Unlock()
	if key_value[slice2[0]] == nil {
		return
	}
	switch slice2[1] {
	case "finish":
		key_value[slice2[0]].success = "true"
		fmt.Printf("==== reqAddr_finish() ====, key = %v success\n", slice2[0])
	}
}

var (
	once bool = false
	onceLock sync.Mutex    // Mutex to sync the active peer set
	SelfID string = ""
	nodeid_value map[string]int = make(map[string]int)
	key_value map[string]*key5 = make(map[string]*key5)
	key_value_lock sync.Mutex    // Mutex to sync the active peer set
)
type key5 struct {
	sync.Mutex    // Mutex to sync the active peer set
	round [6]int//7 = ok
	finish chan string
	success string
	timeAt int64
	future [6]*msgfuture
}

type msgfuture struct {
	msg map[string]string
}
var msg_1300 string = "1300byteee012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"


