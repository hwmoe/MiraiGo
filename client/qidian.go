// 企点协议相关特殊逻辑

package client

import (
	"bytes"
	"fmt"
	"github.com/hwmoe/MiraiGo/binary"
	"github.com/hwmoe/MiraiGo/client/internal/network"
	"github.com/hwmoe/MiraiGo/client/pb/cmd0x3f6"
	"github.com/hwmoe/MiraiGo/client/pb/cmd0x6ff"
	"github.com/hwmoe/MiraiGo/client/pb/msg"
	"github.com/hwmoe/MiraiGo/internal/proto"
	"github.com/hwmoe/MiraiGo/utils"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	decoders["qidianservice.254"] = decodeGetQQFriendGroupListResponse
	decoders["qidianservice.69"] = decodeLoginExtraResponse
	decoders["HttpConn.0x6ff_501"] = decodeConnKeyResponse
}

// getQiDianAddressDetailList 外部联系人列表
func (c *QQClient) getQiDianAddressDetailList() ([]*FriendInfo, error) {
	req := &cmd0x6ff.C519ReqBody{
		SubCmd: proto.Uint32(33),
		CrmCommonHead: &cmd0x6ff.C519CRMMsgHead{
			CrmSubCmd: proto.Uint32(33),
			VerNo:     proto.Uint32(uint32(utils.ConvertSubVersionToInt(c.version().SortVersionName))),
			KfUin:     proto.Uint64(uint64(c.QiDian.MasterUin)),
			LaborUin:  proto.Uint64(uint64(c.Uin)),
			LoginSig: &cmd0x6ff.LoginSig{
				Sig:  c.QiDian.bigDataReqSession.SessionKey,
				Type: proto.Uint32(27),
			},
		},
		GetAddressDetailListReqBody: &cmd0x6ff.GetAddressDetailListReqBody{
			Timestamp:  proto.Uint32(0),
			Timestamp2: proto.Uint64(0),
		},
	}
	//data, _ := proto.Marshal(req)
	//log.Println(hex.EncodeToString(data))
	rspData, err := c.bigDataRequest(0x519, req)
	if err != nil {
		return nil, errors.Wrap(err, "request error")
	}
	rsp := &cmd0x6ff.C519RspBody{}
	//log.Println(len(rspData))
	//log.Println(hex.EncodeToString(rspData))

	if err = proto.Unmarshal(rspData, rsp); err != nil {
		return nil, errors.Wrap(err, "unmarshal error")
	}
	log.Println(rsp.SubCmd)
	log.Println(rsp.GetAddressDetailListRspBody)

	if rsp.GetAddressDetailListRspBody == nil {
		return nil, errors.New("rsp body is nil")
	}
	ret := []*FriendInfo{}
	for _, detail := range rsp.GetAddressDetailListRspBody.AddressDetail {
		if len(detail.Qq) == 0 {
			c.warning("address detail %v QQ is 0", string(detail.Name))
			continue
		}
		ret = append(ret, &FriendInfo{
			Uin:      int64(detail.Qq[0].Account.Unwrap()),
			Nickname: string(detail.Name),
		})
	}
	return ret, nil
}

func (c *QQClient) getQQFriendGroupList() (uint16, []byte) {
	req := &cmd0x3f6.C3F6ReqBody{
		SubCmd: proto.Uint32(254),
		CrmCommonHead: &cmd0x3f6.C3F6CRMMsgHead{
			CrmSubCmd:  proto.Uint32(254),
			KfUin:      proto.Uint64(uint64(c.QiDian.MasterUin)),
			VerNo:      proto.Uint32(uint32(utils.ConvertSubVersionToInt(c.version().SortVersionName))),
			Clienttype: proto.Uint32(2),
			LaborUin:   proto.Uint64(uint64(c.Uin)),
		},
		GetQQFriendGroupListReqBody: &cmd0x3f6.GetQQFriendGroupListReqBody{},
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("qidianservice.254", payload)
}

func (c *QQClient) buildLoginExtraPacket() (uint16, []byte) {
	req := &cmd0x3f6.C3F6ReqBody{
		SubCmd: proto.Uint32(69),
		CrmCommonHead: &cmd0x3f6.C3F6CRMMsgHead{
			CrmSubCmd:  proto.Uint32(69),
			VerNo:      proto.Uint32(uint32(utils.ConvertSubVersionToInt(c.version().SortVersionName))),
			Clienttype: proto.Uint32(2),
		},
		SubcmdLoginProcessCompleteReqBody: &cmd0x3f6.QDUserLoginProcessCompleteReqBody{
			Kfext:        proto.Uint64(uint64(c.Uin)),
			Pubno:        proto.Some(c.version().AppId),
			Buildno:      proto.Uint32(uint32(utils.ConvertSubVersionToInt(c.version().SortVersionName))),
			TerminalType: proto.Uint32(2),
			Status:       proto.Uint32(10),
			LoginTime:    proto.Uint32(5),
			HardwareInfo: proto.String(string(c.Device().Model)),
			SoftwareInfo: proto.String(string(c.Device().Version.Release)),
			Guid:         c.Device().Guid,
			AppName:      proto.Some(c.version().ApkId),
			SubAppId:     proto.Some(c.version().AppId),
		},
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("qidianservice.69", payload)
}

func (c *QQClient) buildConnKeyRequestPacket() (uint16, []byte) {
	req := &cmd0x6ff.C501ReqBody{
		ReqBody: &cmd0x6ff.SubCmd0X501ReqBody{
			Uin:          proto.Uint64(uint64(c.Uin)),
			IdcId:        proto.Uint32(0),
			Appid:        proto.Uint32(16),
			LoginSigType: proto.Uint32(1),
			RequestFlag:  proto.Uint32(3),
			ServiceTypes: []uint32{1},
		},
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("HttpConn.0x6ff_501", payload)
}

func (c *QQClient) bigDataRequest(subCmd uint32, req proto.Message) ([]byte, error) {
	if c.QiDian.bigDataReqSession == nil {
		return nil, errors.New("please call conn key request method before")
	}
	version := strings.Join(strings.Split(c.version().SortVersionName, "."), "")
	var versionNum int
	versionNum, _ = strconv.Atoi(version)
	data, _ := proto.Marshal(req)
	head, _ := proto.Marshal(&msg.IMHead{
		HeadType: proto.Uint32(4),
		HttpconnHead: &msg.HttpConnHead{
			Uin:          proto.Uint64(uint64(c.Uin)),
			Command:      proto.Uint32(1791),
			SubCommand:   proto.Some(subCmd),
			Seq:          proto.Uint32(uint32(c.nextHighwayApplySeq())),
			Version:      proto.Uint32(uint32(versionNum)), // todo: short version convert
			Flag:         proto.Uint32(1),
			CompressType: proto.Uint32(0),
			ErrorCode:    proto.Uint32(0),
		},
		LoginSig: &msg.LoginSig{
			Type: proto.Uint32(22),
			Sig:  c.QiDian.bigDataReqSession.SigSession,
		},
	})
	tea := binary.NewTeaCipher(c.QiDian.bigDataReqSession.SessionKey)
	body := tea.Encrypt(data)
	url := fmt.Sprintf("http://%v/cgi-bin/httpconn", c.QiDian.bigDataReqAddrs[0])
	httpReq, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(binary.NewWriterF(func(w *binary.Writer) {
		w.WriteByte(40)
		w.WriteUInt32(uint32(len(head)))
		w.WriteUInt32(uint32(len(body)))
		w.Write(head)
		w.Write(body)
		w.WriteByte(41)
	})))
	rsp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "request error")
	}
	defer func() { _ = rsp.Body.Close() }()
	rspBody, _ := io.ReadAll(rsp.Body)
	if len(rspBody) == 0 {
		return nil, errors.Wrap(err, "request error")
	}
	r := binary.NewReader(rspBody)
	r.ReadByte()
	l1 := int(r.ReadInt32())
	l2 := int(r.ReadInt32())
	r.ReadBytes(l1)
	payload := r.ReadBytes(l2)
	return tea.Decrypt(payload), nil
}

func decodeLoginExtraResponse(c *QQClient, pkt *network.Packet) (any, error) {
	rsp := cmd0x3f6.C3F6RspBody{}
	if err := proto.Unmarshal(pkt.Payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.SubcmdLoginProcessCompleteRspBody == nil {
		return nil, errors.New("login process resp is nil")
	}
	c.QiDian = &QiDianAccountInfo{
		MasterUin:  int64(rsp.SubcmdLoginProcessCompleteRspBody.Corpuin.Unwrap()),
		ExtName:    rsp.SubcmdLoginProcessCompleteRspBody.ExtuinName.Unwrap(),
		CreateTime: int64(rsp.SubcmdLoginProcessCompleteRspBody.OpenAccountTime.Unwrap()),
	}
	return nil, nil
}

func decodeConnKeyResponse(c *QQClient, pkt *network.Packet) (any, error) {
	rsp := cmd0x6ff.C501RspBody{}
	if err := proto.Unmarshal(pkt.Payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if c.QiDian == nil {
		return nil, errors.New("please call login extra before")
	}
	c.QiDian.bigDataReqSession = &bigDataSessionInfo{
		SigSession: rsp.RspBody.SigSession,
		SessionKey: rsp.RspBody.SessionKey,
	}
	for _, srv := range rsp.RspBody.Addrs {
		if srv.ServiceType.Unwrap() == 1 {
			for _, addr := range srv.Addrs {
				c.QiDian.bigDataReqAddrs = append(c.QiDian.bigDataReqAddrs, fmt.Sprintf("%v:%v", binary.UInt32ToIPV4Address(addr.Ip.Unwrap()), addr.Port.Unwrap()))
			}
		}
	}
	return nil, nil
}

func decodeGetQQFriendGroupListResponse(c *QQClient, pkt *network.Packet) (any, error) {
	rsp := cmd0x3f6.C3F6RspBody{}
	if err := proto.Unmarshal(pkt.Payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.GetQQFriendGroupListRspBody == nil {
		return nil, errors.New("get friend group list resp is nil")
	}

	for _, srv := range rsp.GetQQFriendGroupListRspBody.QQFriendGroupInfoList.QQFriendGroupInfo {
		c.logger.Error(srv.GroupName.Unwrap())
	}

	return nil, nil
}
