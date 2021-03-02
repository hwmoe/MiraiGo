package client

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x352"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x6ff"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/notify"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/client/pb/profilecard"
	"github.com/Mrs4s/MiraiGo/client/pb/qweb"
	"github.com/Mrs4s/MiraiGo/client/pb/structmsg"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

var (
	groupJoinLock  sync.Mutex
	groupLeaveLock sync.Mutex
)

// wtlogin.login
func decodeLoginResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	reader := binary.NewReader(payload)
	reader.ReadUInt16() // sub command
	t := reader.ReadByte()
	reader.ReadUInt16()
	m := reader.ReadTlvMap(2)
	if m.Exists(0x402) {
		c.dpwd = []byte(utils.RandomString(16))
		c.t402 = m[0x402]
		h := md5.Sum(append(append(SystemDeviceInfo.Guid, c.dpwd...), c.t402...))
		c.g = h[:]
	}
	if t == 0 { // login success
		if t150, ok := m[0x150]; ok {
			c.t150 = t150
		}
		if t161, ok := m[0x161]; ok {
			c.decodeT161(t161)
		}
		if m.Exists(0x403) {
			c.randSeed = m[0x403]
		}
		c.decodeT119(m[0x119])
		return LoginResponse{
			Success: true,
		}, nil
	}
	if t == 2 {
		c.t104, _ = m[0x104]
		if m.Exists(0x192) {
			return LoginResponse{
				Success:   false,
				VerifyUrl: string(m[0x192]),
				Error:     SliderNeededError,
			}, nil
		}
		if m.Exists(0x165) { // image
			imgData := binary.NewReader(m[0x105])
			signLen := imgData.ReadUInt16()
			imgData.ReadUInt16()
			sign := imgData.ReadBytes(int(signLen))
			return LoginResponse{
				Success:      false,
				Error:        NeedCaptcha,
				CaptchaImage: imgData.ReadAvailable(),
				CaptchaSign:  sign,
			}, nil
		} else {
			return LoginResponse{
				Success: false,
				Error:   UnknownLoginError,
			}, nil
		}
	} // need captcha

	if t == 40 {
		return LoginResponse{
			Success:      false,
			ErrorMessage: "账号被冻结",
			Error:        UnknownLoginError,
		}, nil
	}

	if t == 160 || t == 239 {
		if t174, ok := m[0x174]; ok { // 短信验证
			c.t104 = m[0x104]
			c.t174 = t174
			c.randSeed = m[0x403]
			phone := func() string {
				r := binary.NewReader(m[0x178])
				return r.ReadStringLimit(int(r.ReadInt32()))
			}()
			if t204, ok := m[0x204]; ok { // 同时支持扫码验证 ?
				return LoginResponse{
					Success:      false,
					Error:        SMSOrVerifyNeededError,
					VerifyUrl:    string(t204),
					SMSPhone:     phone,
					ErrorMessage: string(m[0x17e]),
				}, nil
			}
			return LoginResponse{
				Success:      false,
				Error:        SMSNeededError,
				SMSPhone:     phone,
				ErrorMessage: string(m[0x17e]),
			}, nil
		}

		if _, ok := m[0x17b]; ok { // 二次验证
			c.t104 = m[0x104]
			return LoginResponse{
				Success: false,
				Error:   SMSNeededError,
			}, nil
		}

		if t204, ok := m[0x204]; ok { // 扫码验证
			return LoginResponse{
				Success:      false,
				Error:        UnsafeDeviceError,
				VerifyUrl:    string(t204),
				ErrorMessage: "",
			}, nil
		}

	}

	if t == 162 {
		return LoginResponse{
			Error: TooManySMSRequestError,
		}, nil
	}

	if t == 204 {
		c.t104 = m[0x104]
		c.randSeed = m[0x403]
		return c.sendAndWait(c.buildDeviceLockLoginPacket())
	} // drive lock

	if t149, ok := m[0x149]; ok {
		t149r := binary.NewReader(t149)
		t149r.ReadBytes(2)
		t149r.ReadStringShort() // title
		return LoginResponse{
			Success:      false,
			Error:        OtherLoginError,
			ErrorMessage: t149r.ReadStringShort(),
		}, nil
	}

	if t146, ok := m[0x146]; ok {
		t146r := binary.NewReader(t146)
		t146r.ReadBytes(4)      // ver and code
		t146r.ReadStringShort() // title
		return LoginResponse{
			Success:      false,
			Error:        OtherLoginError,
			ErrorMessage: t146r.ReadStringShort(),
		}, nil
	}
	c.Debug("unknown login response: %v", t)
	for k, v := range m {
		c.Debug("Type: %v Value: %v", strconv.FormatInt(int64(k), 16), hex.EncodeToString(v))
	}
	return nil, errors.Errorf("unknown login response: %v", t) // ?
}

// StatSvc.register
func decodeClientRegisterResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	svcRsp := &jce.SvcRespRegister{}
	svcRsp.ReadFrom(jce.NewJceReader(data.Map["SvcRespRegister"]["QQService.SvcRespRegister"][1:]))
	if svcRsp.Result != "" || svcRsp.ReplyCode != 0 {
		if svcRsp.Result != "" {
			c.Error("reg error: %v", svcRsp.Result)
		}
		return nil, errors.New("reg failed")
	}
	return nil, nil
}

// wtlogin.exchange_emp
func decodeExchangeEmpResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	reader := binary.NewReader(payload)
	cmd := reader.ReadUInt16()
	t := reader.ReadByte()
	reader.ReadUInt16()
	m := reader.ReadTlvMap(2)
	if t != 0 {
		c.Error("exchange_emp error: %v", t)
		return nil, nil
	}
	if cmd == 15 { // TODO: 免密登录
		c.decodeT119R(m[0x119])
	}
	return nil, nil
}

// ConfigPushSvc.PushReq
func decodePushReqPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["PushReq"]["ConfigPush.PushReq"][1:])
	t := r.ReadInt32(1)
	jceBuf := r.ReadAny(2).([]byte)
	if len(jceBuf) > 0 {
		switch t {
		case 1:
			ssoPkt := jce.NewJceReader(jceBuf)
			servers := []jce.SsoServerInfo{}
			ssoPkt.ReadSlice(&servers, 1)
			if len(servers) > 0 {
				var adds []*net.TCPAddr
				for _, s := range servers {
					if strings.Contains(s.Server, "com") {
						continue
					}
					c.Debug("got new server addr: %v location: %v", s.Server, s.Location)
					adds = append(adds, &net.TCPAddr{
						IP:   net.ParseIP(s.Server),
						Port: int(s.Port),
					})
				}
				f := true
				for _, e := range c.eventHandlers.serverUpdatedHandlers {
					cover(func() {
						if !e(c, &ServerUpdatedEvent{Servers: servers}) {
							f = false
						}
					})
				}
				if f {
					c.SetCustomServer(adds)
				}
				return nil, nil
			}
		case 2:
			fmtPkt := jce.NewJceReader(jceBuf)
			list := &jce.FileStoragePushFSSvcList{}
			list.ReadFrom(fmtPkt)
			c.Debug("got file storage svc push.")
			c.fileStorageInfo = list
			rsp := cmd0x6ff.C501RspBody{}
			if err := proto.Unmarshal(list.BigDataChannel.PbBuf, &rsp); err == nil && rsp.RspBody != nil {
				c.highwaySession = &highwaySessionInfo{
					SigSession: rsp.RspBody.SigSession,
					SessionKey: rsp.RspBody.SessionKey,
				}
				for _, srv := range rsp.RspBody.Addrs {
					if srv.GetServiceType() == 10 {
						for _, addr := range srv.Addrs {
							c.srvSsoAddrs = append(c.srvSsoAddrs, fmt.Sprintf("%v:%v", binary.UInt32ToIPV4Address(addr.GetIp()), addr.GetPort()))
						}
					}
					if srv.GetServiceType() == 21 {
						for _, addr := range srv.Addrs {
							c.otherSrvAddrs = append(c.otherSrvAddrs, fmt.Sprintf("%v:%v", binary.UInt32ToIPV4Address(addr.GetIp()), addr.GetPort()))
						}
					}
				}
			}
		}
	}

	seq := r.ReadInt64(3)
	_, pkt := c.buildConfPushRespPacket(t, seq, jceBuf)
	return nil, c.send(pkt)
}

// MessageSvc.PbGetMsg
func decodeMessageSvcPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := msg.GetMessageResponse{}
	err := proto.Unmarshal(payload, &rsp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	c.c2cMessageSyncProcessor(&rsp, &c2cExtraOption{UsedRegProxy: false})
	return nil, nil
}

// MessageSvc.PushNotify
func decodeSvcNotify(c *QQClient, _ uint16, _ []byte) (interface{}, error) {
	_, err := c.sendAndWait(c.buildGetMessageRequestPacket(msg.SyncFlag_START, time.Now().Unix()))
	return nil, err
}

// SummaryCard.ReqSummaryCard
func decodeSummaryCardResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	rsp := func() *jce.JceReader {
		if r, ok := data.Map["RespSummaryCard"]["SummaryCard.RespSummaryCard"]; ok {
			return jce.NewJceReader(r[1:])
		}
		return jce.NewJceReader(data.Map["RespSummaryCard"]["SummaryCard_Old.RespSummaryCard"][1:])
	}()
	info := &SummaryCardInfo{
		Sex:       rsp.ReadByte(1),
		Age:       rsp.ReadByte(2),
		Nickname:  rsp.ReadString(3),
		Level:     rsp.ReadInt32(5),
		City:      rsp.ReadString(7),
		Sign:      rsp.ReadString(8),
		Mobile:    rsp.ReadString(11),
		Uin:       rsp.ReadInt64(23),
		LoginDays: rsp.ReadInt64(36),
	}
	services := [][]byte{}
	rsp.ReadSlice(&services, 46)
	readService := func(buf []byte) (*profilecard.BusiComm, []byte) {
		r := binary.NewReader(buf)
		r.ReadByte()
		l1 := r.ReadInt32()
		l2 := r.ReadInt32()
		comm := r.ReadBytes(int(l1))
		d := r.ReadBytes(int(l2))
		c := &profilecard.BusiComm{}
		_ = proto.Unmarshal(comm, c)
		return c, d
	}
	for _, buf := range services {
		comm, payload := readService(buf)
		if comm.GetService() == 16 {
			rsp := profilecard.GateVaProfileGateRsp{}
			_ = proto.Unmarshal(payload, &rsp)
			if rsp.QidInfo != nil {
				info.Qid = rsp.QidInfo.GetQid()
			}
		}
	}
	return info, nil
}

// friendlist.getFriendGroupList
func decodeFriendGroupListResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["FLRESP"][1:])
	totalFriendCount := r.ReadInt16(5)
	friends := []jce.FriendInfo{}
	r.ReadSlice(&friends, 7)
	var l []*FriendInfo
	for _, f := range friends {
		l = append(l, &FriendInfo{
			Uin:      f.FriendUin,
			Nickname: f.Nick,
			Remark:   f.Remark,
			FaceId:   f.FaceId,
		})
	}
	rsp := FriendListResponse{
		TotalCount: int32(totalFriendCount),
		List:       l,
	}
	return rsp, nil
}

// friendlist.GetTroopListReqV2
func decodeGroupListResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["GetTroopListRespV2"][1:])
	vecCookie := []byte{}
	groups := []jce.TroopNumber{}
	r.ReadSlice(&vecCookie, 4)
	r.ReadSlice(&groups, 5)
	var l []*GroupInfo
	for _, g := range groups {
		l = append(l, &GroupInfo{
			Uin:            g.GroupUin,
			Code:           g.GroupCode,
			Name:           g.GroupName,
			Memo:           g.GroupMemo,
			OwnerUin:       g.GroupOwnerUin,
			MemberCount:    uint16(g.MemberNum),
			MaxMemberCount: uint16(g.MaxGroupMemberNum),
			client:         c,
		})
	}
	if len(vecCookie) > 0 {
		rsp, err := c.sendAndWait(c.buildGroupListRequestPacket(vecCookie))
		if err != nil {
			return nil, err
		}
		l = append(l, rsp.([]*GroupInfo)...)
	}
	return l, nil
}

// friendlist.GetTroopMemberListReq
func decodeGroupMemberListResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["GTMLRESP"][1:])
	members := []jce.TroopMemberInfo{}
	r.ReadSlice(&members, 3)
	next := r.ReadInt64(4)
	var l []*GroupMemberInfo
	for _, m := range members {
		l = append(l, &GroupMemberInfo{
			Uin:                    m.MemberUin,
			Nickname:               m.Nick,
			Gender:                 m.Gender,
			CardName:               m.Name,
			Level:                  uint16(m.MemberLevel),
			JoinTime:               m.JoinTime,
			LastSpeakTime:          m.LastSpeakTime,
			SpecialTitle:           m.SpecialTitle,
			SpecialTitleExpireTime: m.SpecialTitleExpireTime,
			Permission: func() MemberPermission {
				if m.Flag == 1 {
					return Administrator
				}
				return Member
			}(),
		})
	}
	return groupMemberListResponse{
		NextUin: next,
		list:    l,
	}, nil
}

// group_member_card.get_group_member_card_info
func decodeGroupMemberInfoResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := pb.GroupMemberRspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.MemInfo == nil || (rsp.MemInfo.Nick == nil && rsp.MemInfo.Age == 0) {
		return nil, errors.WithStack(ErrMemberNotFound)
	}
	group := c.FindGroup(rsp.GroupCode)
	return &GroupMemberInfo{
		Group:                  group,
		Uin:                    rsp.MemInfo.Uin,
		Gender:                 byte(rsp.MemInfo.Sex),
		Nickname:               string(rsp.MemInfo.Nick),
		CardName:               string(rsp.MemInfo.Card),
		Level:                  uint16(rsp.MemInfo.Level),
		JoinTime:               rsp.MemInfo.Join,
		LastSpeakTime:          rsp.MemInfo.LastSpeak,
		SpecialTitle:           string(rsp.MemInfo.SpecialTitle),
		SpecialTitleExpireTime: int64(rsp.MemInfo.SpecialTitleExpireTime),
		Permission: func() MemberPermission {
			if rsp.MemInfo.Uin == group.OwnerUin {
				return Owner
			}
			if rsp.MemInfo.Role == 2 {
				return Administrator
			}
			return Member
		}(),
	}, nil
}

// LongConn.OffPicUp
func decodeOffPicUpResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := cmd0x352.RspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.FailMsg != "" {
		return imageUploadResponse{
			ResultCode: -1,
			Message:    rsp.FailMsg,
		}, nil
	}
	if rsp.Subcmd != 1 || len(rsp.MsgTryupImgRsp) == 0 {
		return imageUploadResponse{
			ResultCode: -2,
		}, nil
	}
	imgRsp := rsp.MsgTryupImgRsp[0]
	if imgRsp.Result != 0 {
		return imageUploadResponse{
			ResultCode: imgRsp.Result,
			Message:    imgRsp.FailMsg,
		}, nil
	}
	if imgRsp.BoolFileExit {
		return imageUploadResponse{
			IsExists:   true,
			ResourceId: imgRsp.UpResid,
		}, nil
	}
	return imageUploadResponse{
		ResourceId: imgRsp.UpResid,
		UploadKey:  imgRsp.UpUkey,
		UploadIp:   imgRsp.Uint32UpIp,
		UploadPort: imgRsp.Uint32UpPort,
	}, nil
}

// OnlinePush.ReqPush
func decodeOnlinePushReqPacket(c *QQClient, seq uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	jr := jce.NewJceReader(data.Map["req"]["OnlinePushPack.SvcReqPushMsg"][1:])
	msgInfos := []jce.PushMessageInfo{}
	uin := jr.ReadInt64(0)
	jr.ReadSlice(&msgInfos, 2)
	_ = c.send(c.buildDeleteOnlinePushPacket(uin, seq, msgInfos))
	for _, m := range msgInfos {
		k := fmt.Sprintf("%v%v%v", m.MsgSeq, m.MsgTime, m.MsgUid)
		if _, ok := c.onlinePushCache.Get(k); ok {
			continue
		}
		c.onlinePushCache.Add(k, "", time.Second*30)
		if m.MsgType == 732 {
			r := binary.NewReader(m.VMsg)
			groupId := int64(uint32(r.ReadInt32()))
			iType := r.ReadByte()
			r.ReadByte()
			switch iType {
			case 0x0c: // 群内禁言
				operator := int64(uint32(r.ReadInt32()))
				if operator == c.Uin {
					continue
				}
				r.ReadBytes(6)
				target := int64(uint32(r.ReadInt32()))
				t := r.ReadInt32()
				c.dispatchGroupMuteEvent(&GroupMuteEvent{
					GroupCode:   groupId,
					OperatorUin: operator,
					TargetUin:   target,
					Time:        t,
				})
			case 0x10, 0x11, 0x14, 0x15: // group notify msg
				r.ReadByte()
				b := notify.NotifyMsgBody{}
				_ = proto.Unmarshal(r.ReadAvailable(), &b)
				if b.OptMsgRecall != nil {
					for _, rm := range b.OptMsgRecall.RecalledMsgList {
						if rm.MsgType == 2 {
							continue
						}
						c.dispatchGroupMessageRecalledEvent(&GroupMessageRecalledEvent{
							GroupCode:   groupId,
							OperatorUin: b.OptMsgRecall.Uin,
							AuthorUin:   rm.AuthorUin,
							MessageId:   rm.Seq,
							Time:        rm.Time,
						})
					}
				}
				if b.OptGeneralGrayTip != nil {
					c.grayTipProcessor(groupId, b.OptGeneralGrayTip)
				}
				if b.OptMsgRedTips != nil {
					if b.OptMsgRedTips.LuckyFlag == 1 { // 运气王提示
						c.dispatchGroupNotifyEvent(&GroupRedBagLuckyKingNotifyEvent{
							GroupCode: groupId,
							Sender:    int64(b.OptMsgRedTips.SenderUin),
							LuckyKing: int64(b.OptMsgRedTips.LuckyUin),
						})
					}
				}
				if b.QqGroupDigestMsg != nil {
					var digest = b.QqGroupDigestMsg
					c.dispatchGroupDigestEvent(&GroupDigestEvent{
						GroupCode:         int64(digest.GroupCode),
						MessageID:         int32(digest.Seq),
						InternalMessageID: int32(digest.Random),
						OperationType:     digest.OpType,
						OperateTime:       digest.OpTime,
						SenderUin:         int64(digest.Sender),
						OperatorUin:       int64(digest.DigestOper),
						SenderNick:        string(digest.SenderNick),
						OperatorNick:      string(digest.OperNick),
					})
				}
			}
		}
		if m.MsgType == 528 {
			vr := jce.NewJceReader(m.VMsg)
			subType := vr.ReadInt64(0)
			probuf := vr.ReadAny(10).([]byte)
			switch subType {
			case 0x8A, 0x8B:
				s8a := pb.Sub8A{}
				if err := proto.Unmarshal(probuf, &s8a); err != nil {
					return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
				}
				for _, m := range s8a.MsgInfo {
					if m.ToUin == c.Uin {
						c.dispatchFriendMessageRecalledEvent(&FriendMessageRecalledEvent{
							FriendUin: m.FromUin,
							MessageId: m.MsgSeq,
							Time:      m.MsgTime,
						})
					}
				}
			case 0xB3:
				b3 := pb.SubB3{}
				if err := proto.Unmarshal(probuf, &b3); err != nil {
					return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
				}
				frd := &FriendInfo{
					Uin:      b3.MsgAddFrdNotify.Uin,
					Nickname: b3.MsgAddFrdNotify.Nick,
				}
				c.FriendList = append(c.FriendList, frd)
				c.dispatchNewFriendEvent(&NewFriendEvent{Friend: frd})
			case 0xD4:
				d4 := pb.SubD4{}
				if err := proto.Unmarshal(probuf, &d4); err != nil {
					return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
				}
				groupLeaveLock.Lock()
				if g := c.FindGroup(d4.Uin); g != nil {
					if err := c.ReloadGroupList(); err != nil {
						groupLeaveLock.Unlock()
						return nil, err
					}
					c.dispatchLeaveGroupEvent(&GroupLeaveEvent{Group: g})
				}
				groupLeaveLock.Unlock()
			case 0x27:
				s27 := pb.Sub27{}
				if err := proto.Unmarshal(probuf, &s27); err != nil {
					return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
				}
				for _, m := range s27.ModInfos {
					if m.DelFriend != nil {
						frdUin := m.DelFriend.Uins[0]
						if frd := c.FindFriend(int64(frdUin)); frd != nil {
							log.Infof("好友被删除: %v(%v)", frd.Nickname, frd.Uin)
							if err := c.ReloadFriendList(); err != nil {
								return nil, errors.Wrap(err, "failed to reload friend list")
							}
						}
					}
				}
			case 290:
				t := &notify.GeneralGrayTipInfo{}
				_ = proto.Unmarshal(probuf, t)
				var sender int64
				for _, templ := range t.MsgTemplParam {
					if templ.Name == "uin_str1" {
						sender, _ = strconv.ParseInt(templ.Value, 10, 64)
					}
				}
				if sender == 0 {
					return nil, nil
				}
				c.dispatchFriendNotifyEvent(&FriendPokeNotifyEvent{
					Sender:   sender,
					Receiver: c.Uin,
				})
			case 0x44:
				s44 := pb.Sub44{}
				if err := proto.Unmarshal(probuf, &s44); err != nil {
					return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
				}
				if s44.GroupSyncMsg != nil {
					func() {
						groupJoinLock.Lock()
						defer groupJoinLock.Unlock()
						if s44.GroupSyncMsg.GetGrpCode() != 0 { // member sync
							c.Debug("syncing members.")
							if group := c.FindGroup(s44.GroupSyncMsg.GetGrpCode()); group != nil {
								group.Update(func(_ *GroupInfo) {
									var lastJoinTime int64 = 0
									for _, m := range group.Members {
										if lastJoinTime < m.JoinTime {
											lastJoinTime = m.JoinTime
										}
									}
									if newMem, err := c.GetGroupMembers(group); err == nil {
										group.Members = newMem
										for _, m := range newMem {
											if lastJoinTime < m.JoinTime {
												go c.dispatchNewMemberEvent(&MemberJoinGroupEvent{
													Group:  group,
													Member: m,
												})
											}
										}
									}
								})
							}
						}
					}()
				}
			}
		}
	}
	return nil, nil
}

// OnlinePush.PbPushTransMsg
func decodeOnlinePushTransPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	info := msg.TransMsgInfo{}
	err := proto.Unmarshal(payload, &info)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	data := binary.NewReader(info.MsgData)
	idStr := strconv.FormatInt(info.GetMsgUid(), 10)
	if _, ok := c.transCache.Get(idStr); ok {
		return nil, nil
	}
	c.transCache.Add(idStr, "", time.Second*15)
	if info.GetMsgType() == 34 {
		data.ReadInt32()
		data.ReadByte()
		target := int64(uint32(data.ReadInt32()))
		typ := int32(data.ReadByte())
		operator := int64(uint32(data.ReadInt32()))
		if g := c.FindGroupByUin(info.GetFromUin()); g != nil {
			groupLeaveLock.Lock()
			defer groupLeaveLock.Unlock()
			switch typ {
			case 0x02:
				if target == c.Uin {
					c.dispatchLeaveGroupEvent(&GroupLeaveEvent{Group: g})
				} else {
					if m := g.FindMember(target); m != nil {
						g.removeMember(target)
						c.dispatchMemberLeaveEvent(&MemberLeaveGroupEvent{
							Group:  g,
							Member: m,
						})
					}
				}
			case 0x03:
				if err = c.ReloadGroupList(); err != nil {
					return nil, err
				}
				if target == c.Uin {
					c.dispatchLeaveGroupEvent(&GroupLeaveEvent{
						Group:    g,
						Operator: g.FindMember(operator),
					})
				} else {
					if m := g.FindMember(target); m != nil {
						g.removeMember(target)
						c.dispatchMemberLeaveEvent(&MemberLeaveGroupEvent{
							Group:    g,
							Member:   m,
							Operator: g.FindMember(operator),
						})
					}
				}
			case 0x82:
				if m := g.FindMember(target); m != nil {
					g.removeMember(target)
					c.dispatchMemberLeaveEvent(&MemberLeaveGroupEvent{
						Group:  g,
						Member: m,
					})
				}
			case 0x83:
				if m := g.FindMember(target); m != nil {
					g.removeMember(target)
					c.dispatchMemberLeaveEvent(&MemberLeaveGroupEvent{
						Group:    g,
						Member:   m,
						Operator: g.FindMember(operator),
					})
				}
			}
		}
	}
	if info.GetMsgType() == 44 {
		data.ReadBytes(5)
		var4 := int32(data.ReadByte())
		var var5 int64 = 0
		target := int64(uint32(data.ReadInt32()))
		if var4 != 0 && var4 != 1 {
			var5 = int64(uint32(data.ReadInt32()))
		}
		if g := c.FindGroupByUin(info.GetFromUin()); g != nil {
			if var5 == 0 && data.Len() == 1 {
				newPermission := func() MemberPermission {
					if data.ReadByte() == 1 {
						return Administrator
					}
					return Member
				}()
				mem := g.FindMember(target)
				if mem.Permission != newPermission {
					old := mem.Permission
					mem.Permission = newPermission
					c.dispatchPermissionChanged(&MemberPermissionChangedEvent{
						Group:         g,
						Member:        mem,
						OldPermission: old,
						NewPermission: newPermission,
					})
				}
			}
		}
	}
	return nil, nil
}

// ProfileService.Pb.ReqSystemMsgNew.Friend
func decodeSystemMsgFriendPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := structmsg.RspSystemMsgNew{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(rsp.Friendmsgs) == 0 {
		return nil, nil
	}
	st := rsp.Friendmsgs[0]
	if st.Msg != nil {
		c.dispatchNewFriendRequest(&NewFriendRequest{
			RequestId:     st.MsgSeq,
			Message:       st.Msg.MsgAdditional,
			RequesterUin:  st.ReqUin,
			RequesterNick: st.Msg.ReqUinNick,
			client:        c,
		})
	}
	return nil, nil
}

// MessageSvc.PushForceOffline
func decodeForceOfflinePacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["req_PushForceOffline"]["PushNotifyPack.RequestPushForceOffline"][1:])
	tips := r.ReadString(2)
	c.lastLostMsg = tips
	c.NetLooping = false
	c.Online = false
	return nil, nil
}

// StatSvc.ReqMSFOffline
func decodeMSFOfflinePacket(c *QQClient, _ uint16, _ []byte) (interface{}, error) {
	c.lastLostMsg = "服务器端强制下线."
	c.NetLooping = false
	c.Online = false
	return nil, nil
}

// OidbSvc.0xd79
func decodeWordSegmentation(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := &oidb.D79RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.Content != nil {
		return rsp.Content.SliceContent, nil
	}
	return nil, errors.New("no word received")
}

// OidbSvc.0xe07_0
func decodeImageOcrResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.DE07RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.Wording != "" {
		return nil, errors.New(rsp.Wording)
	}
	if rsp.RetCode != 0 {
		return nil, errors.New(fmt.Sprintf("server error, code: %v msg: %v", rsp.RetCode, rsp.ErrMsg))
	}
	var texts []*TextDetection
	for _, text := range rsp.OcrRspBody.TextDetections {
		var points []*Coordinate
		for _, c := range text.Polygon.Coordinates {
			points = append(points, &Coordinate{
				X: c.X,
				Y: c.Y,
			})
		}
		texts = append(texts, &TextDetection{
			Text:        text.DetectedText,
			Confidence:  text.Confidence,
			Coordinates: points,
		})
	}
	return &OcrResponse{
		Texts:    texts,
		Language: rsp.OcrRspBody.Language,
	}, nil
}

// LightAppSvc.mini_app_info.GetAppInfoById
func decodeAppInfoResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := qweb.QWebRsp{}
	rsp := qweb.GetAppInfoByIdRsp{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if pkg.RetCode != 0 {
		return nil, errors.New(pkg.ErrMsg)
	}
	if err := proto.Unmarshal(pkg.BusiBuff, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return rsp.AppInfo, nil
}

func ignoreDecoder(_ *QQClient, _ uint16, _ []byte) (interface{}, error) {
	return nil, nil
}
