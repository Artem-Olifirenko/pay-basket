package basket_item

import "sync"

type ConfiguratorItemAdditions struct {
	ConfId   string       // 1
	ConfType ConfType     // 2
	mx       sync.RWMutex `msgpack:"-"`
}

func NewConfiguratorItemAdditions(
	confId string,
	confType ConfType,
) *ConfiguratorItemAdditions {
	return &ConfiguratorItemAdditions{ConfId: confId, ConfType: confType, mx: sync.RWMutex{}}
}

func (c *ConfiguratorItemAdditions) IsMutable() bool {
	if c == nil {
		return false
	}

	confType := c.GetConfType()
	if confType == ConfTypeTemplate || confType == ConfTypeVendor {
		return false
	}

	return true
}

func (c *ConfiguratorItemAdditions) GetConfId() string {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.ConfId
}

func (c *ConfiguratorItemAdditions) SetConfId(v string) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.ConfId = v
}

func (c *ConfiguratorItemAdditions) GetConfType() ConfType {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.ConfType
}

func (c *ConfiguratorItemAdditions) SetConfType(v ConfType) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.ConfType = v
}
