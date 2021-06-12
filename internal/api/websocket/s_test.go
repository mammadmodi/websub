package websocket

import (
	"github.com/golang/mock/gomock"
	"github.com/mammadmodi/websub/pkg/hub"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestNewSockHub(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := Configuration{}
	h := hub.NewMockHub(ctrl)
	l := logrus.New()
	l.SetOutput(ioutil.Discard)
	sh := NewSockHub(c, h, l)
	assert.Equal(t, h, sh.Hub)
	assert.Equal(t, c, sh.Config)
	assert.Equal(t, l, sh.logger)
	assert.NotNil(t, sh.upgrader)
}
