package gorilla

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/go-gotop/kit/websocket"
	mock_websocket "github.com/go-gotop/kit/websocket/mock"
)

func TestSuite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mws := mock_websocket.NewMockWebSocketConn(ctrl)

	w := &websocketTestSuite{
		mws: mws,
		ws:  NewGorillaWebsocket(mws, &websocket.WebsocketConfig{}),
	}
	suite.Run(t, w)
}

type websocketTestSuite struct {
	suite.Suite
	mws *mock_websocket.MockWebSocketConn
	ws  *GorillaWebsocket
}

func (w *websocketTestSuite) TestConnect() {
	// 使用一个channel来同步测试的结束
	done := make(chan bool, 1)

	w.mws.EXPECT().Dial(gomock.Any(), gomock.Any()).Return(nil)

	w.mws.EXPECT().ReadMessage().DoAndReturn(func() (int, []byte, error) {
		done <- true // 在预期的最后一次调用中发送结束信号
		return 1, []byte("message 3"), nil
	}).AnyTimes() // 明确指定调用次数

	// 调用Connect方法
	w.ws.Connect(&websocket.WebsocketRequest{
		Endpoint: "test",
		ID:       "test",
		MessageHandler: func(message []byte) {
			return
		},
		ErrorHandler: func(err error) {},
	})

	<-done // 等待结束信号
	w.Assert().Equal(uint64(1), w.ws.messageCount)
	w.Assert().True(w.ws.isConnected)
	w.Assert().NotZero(w.ws.connectTime)
	w.Assert().NotNil(w.ws.req)
}

func (w *websocketTestSuite) TestDisconnect() {
	w.mws.EXPECT().Close().Return(nil)
	w.ws.Disconnect()
	w.Assert().False(w.ws.isConnected)
}

func (w *websocketTestSuite) TestReconnect() {
	// test case
}

func (w *websocketTestSuite) TestIsConnected() {
	// test case
}

func (w *websocketTestSuite) TestGetCurrentRate() {
	// test case
}

func (w *websocketTestSuite) TestConnectionDuration() {
	// test case
}

func (w *websocketTestSuite) SetupTest() {
	// setup test
}

func (w *websocketTestSuite) TearDownTest() {
	// teardown test
}
