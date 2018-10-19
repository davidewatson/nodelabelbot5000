package nodelabelbot5000

import (
	"github.com/juju/loggo"
	"github.com/samsung-cnct/nodelabelbot5000/pkg/util"
)

var (
	logger loggo.Logger
)

func (c *NodeLabelBot5000Controller) SetLogger() {
	logger = util.GetModuleLogger("pkg.controllers.nodelabelbot5000", loggo.INFO)
}
