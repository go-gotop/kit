// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/websocket/websocket.go
//
// Generated by this command:
//
//	mockgen -source=pkg/websocket/websocket.go -destination=pkg/websocket/mock/websocket_mock.go
//

// Package mock_websocket is a generated GoMock package.
package mock_websocket

import (
	websocket "github.com/go-gotop/kit/websocket"
	http "net/http"
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

// MockWebSocketConn is a mock of WebSocketConn interface.
type MockWebSocketConn struct {
	ctrl     *gomock.Controller
	recorder *MockWebSocketConnMockRecorder
}

// MockWebSocketConnMockRecorder is the mock recorder for MockWebSocketConn.
type MockWebSocketConnMockRecorder struct {
	mock *MockWebSocketConn
}

// NewMockWebSocketConn creates a new mock instance.
func NewMockWebSocketConn(ctrl *gomock.Controller) *MockWebSocketConn {
	mock := &MockWebSocketConn{ctrl: ctrl}
	mock.recorder = &MockWebSocketConnMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWebSocketConn) EXPECT() *MockWebSocketConnMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockWebSocketConn) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockWebSocketConnMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockWebSocketConn)(nil).Close))
}

// Dial mocks base method.
func (m *MockWebSocketConn) Dial(endpoint string, requestHeader http.Header) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dial", endpoint, requestHeader)
	ret0, _ := ret[0].(error)
	return ret0
}

// Dial indicates an expected call of Dial.
func (mr *MockWebSocketConnMockRecorder) Dial(endpoint, requestHeader any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dial", reflect.TypeOf((*MockWebSocketConn)(nil).Dial), endpoint, requestHeader)
}

// ReadMessage mocks base method.
func (m *MockWebSocketConn) ReadMessage() (int, []byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadMessage")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ReadMessage indicates an expected call of ReadMessage.
func (mr *MockWebSocketConnMockRecorder) ReadMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadMessage", reflect.TypeOf((*MockWebSocketConn)(nil).ReadMessage))
}

// SetPingHandler mocks base method.
func (m *MockWebSocketConn) SetPingHandler(h func(string) error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetPingHandler", h)
}

// SetPingHandler indicates an expected call of SetPingHandler.
func (mr *MockWebSocketConnMockRecorder) SetPingHandler(h any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPingHandler", reflect.TypeOf((*MockWebSocketConn)(nil).SetPingHandler), h)
}

// SetPongHandler mocks base method.
func (m *MockWebSocketConn) SetPongHandler(h func(string) error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetPongHandler", h)
}

// SetPongHandler indicates an expected call of SetPongHandler.
func (mr *MockWebSocketConnMockRecorder) SetPongHandler(h any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPongHandler", reflect.TypeOf((*MockWebSocketConn)(nil).SetPongHandler), h)
}

// WriteMessage mocks base method.
func (m *MockWebSocketConn) WriteMessage(messageType int, data []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteMessage", messageType, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteMessage indicates an expected call of WriteMessage.
func (mr *MockWebSocketConnMockRecorder) WriteMessage(messageType, data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteMessage", reflect.TypeOf((*MockWebSocketConn)(nil).WriteMessage), messageType, data)
}

// MockWebsocket is a mock of Websocket interface.
type MockWebsocket struct {
	ctrl     *gomock.Controller
	recorder *MockWebsocketMockRecorder
}

// MockWebsocketMockRecorder is the mock recorder for MockWebsocket.
type MockWebsocketMockRecorder struct {
	mock *MockWebsocket
}

// NewMockWebsocket creates a new mock instance.
func NewMockWebsocket(ctrl *gomock.Controller) *MockWebsocket {
	mock := &MockWebsocket{ctrl: ctrl}
	mock.recorder = &MockWebsocketMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWebsocket) EXPECT() *MockWebsocketMockRecorder {
	return m.recorder
}

// Connect mocks base method.
func (m *MockWebsocket) Connect(req *websocket.WebsocketRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect", req)
	ret0, _ := ret[0].(error)
	return ret0
}

// Connect indicates an expected call of Connect.
func (mr *MockWebsocketMockRecorder) Connect(req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockWebsocket)(nil).Connect), req)
}

// ConnectionDuration mocks base method.
func (m *MockWebsocket) ConnectionDuration() time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectionDuration")
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// ConnectionDuration indicates an expected call of ConnectionDuration.
func (mr *MockWebsocketMockRecorder) ConnectionDuration() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectionDuration", reflect.TypeOf((*MockWebsocket)(nil).ConnectionDuration))
}

// Disconnect mocks base method.
func (m *MockWebsocket) Disconnect() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Disconnect")
	ret0, _ := ret[0].(error)
	return ret0
}

// Disconnect indicates an expected call of Disconnect.
func (mr *MockWebsocketMockRecorder) Disconnect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Disconnect", reflect.TypeOf((*MockWebsocket)(nil).Disconnect))
}

// GetCurrentRate mocks base method.
func (m *MockWebsocket) GetCurrentRate() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentRate")
	ret0, _ := ret[0].(int)
	return ret0
}

// GetCurrentRate indicates an expected call of GetCurrentRate.
func (mr *MockWebsocketMockRecorder) GetCurrentRate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentRate", reflect.TypeOf((*MockWebsocket)(nil).GetCurrentRate))
}

// IsConnected mocks base method.
func (m *MockWebsocket) IsConnected() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsConnected")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsConnected indicates an expected call of IsConnected.
func (mr *MockWebsocketMockRecorder) IsConnected() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsConnected", reflect.TypeOf((*MockWebsocket)(nil).IsConnected))
}

// Reconnect mocks base method.
func (m *MockWebsocket) Reconnect() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconnect")
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconnect indicates an expected call of Reconnect.
func (mr *MockWebsocketMockRecorder) Reconnect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconnect", reflect.TypeOf((*MockWebsocket)(nil).Reconnect))
}
