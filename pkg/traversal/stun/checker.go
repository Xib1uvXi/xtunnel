package stun

import (
	"context"
)

type Observer interface {
	OnMappingBehaviorChanged(NatBehavior)
	OnFilteringBehaviorChanged(NatBehavior)
	OnNatTypeChanged(NatType)
}

type Checker struct {
	ctx        context.Context
	cancel     context.CancelFunc
	stunServer string

	mappingBehavior   NatBehavior
	filteringBehavior NatBehavior
	natTpye           NatType

	mappingBehaviorChan   chan NatBehavior
	filteringBehaviorChan chan NatBehavior
	natTypeChan           chan NatType

	observers []Observer
}

func NewChecker(ctx context.Context, stunServer string, enableNatCheck bool, obser ...Observer) *Checker {
	ctx, cancel := context.WithCancel(ctx)

	chk := &Checker{
		ctx:                   ctx,
		cancel:                cancel,
		stunServer:            stunServer,
		mappingBehavior:       NatBehavior_NAT_BEHAVIOR_UNKNOWN,
		filteringBehavior:     NatBehavior_NAT_BEHAVIOR_UNKNOWN,
		natTpye:               NatType_NAT_TYPE_UNKNOWN,
		mappingBehaviorChan:   make(chan NatBehavior),
		filteringBehaviorChan: make(chan NatBehavior),
		natTypeChan:           make(chan NatType),
	}

	if enableNatCheck {
		if len(obser) > 0 {
			chk.observers = append(chk.observers, obser...)
		}

		go chk.loop()
		chk.Check()
	}

	return chk
}

func (c *Checker) AddObserver(observer Observer) {
	c.observers = append(c.observers, observer)
}

func (c *Checker) loop() {
	for {
		select {
		case <-c.ctx.Done():
			return

		case mappingBehavior := <-c.mappingBehaviorChan:
			c.mappingBehavior = mappingBehavior
			go func() {
				for _, observer := range c.observers {
					observer.OnMappingBehaviorChanged(mappingBehavior)
				}
			}()

		case filteringBehavior := <-c.filteringBehaviorChan:
			c.filteringBehavior = filteringBehavior
			go func() {
				for _, observer := range c.observers {
					observer.OnFilteringBehaviorChanged(filteringBehavior)
				}
			}()

		case natType := <-c.natTypeChan:
			c.natTpye = natType
			go func() {
				for _, observer := range c.observers {
					observer.OnNatTypeChanged(natType)
				}
			}()
		}
	}
}

func (c *Checker) Check() {
	go func() {
		mappingBehavior, err := MappingTests(c.ctx, c.stunServer)
		if err != nil {
			return
		}
		c.mappingBehaviorChan <- mappingBehavior
	}()

	go func() {
		filteringBehavior, err := FilteringTests(c.ctx, c.stunServer)
		if err != nil {
			return
		}
		c.filteringBehaviorChan <- filteringBehavior
	}()

	go func() {
		natType, err := NATType(c.stunServer)
		if err != nil {
			return
		}
		c.natTypeChan <- natType
	}()
}

func (c *Checker) MappingBehavior() NatBehavior {
	return c.mappingBehavior
}

func (c *Checker) FilteringBehavior() NatBehavior {
	return c.filteringBehavior
}

func (c *Checker) NatType() NatType {
	return c.natTpye
}

// new address
func (c *Checker) NewAddress() (*StunAddressResolver, error) {
	as, err := NewStunAddressResolver(c.ctx, c.stunServer)
	if err != nil {
		return nil, err
	}

	if err := as.Resolve(); err != nil {
		return nil, err
	}

	return as, nil
}

func (c *Checker) Close() {
	c.cancel()
}
