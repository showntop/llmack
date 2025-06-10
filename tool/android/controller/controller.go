package controller

import "context"

type Controller struct {
	registry *Registry
}

func NewController() *Controller {
	ctrl := &Controller{
		registry: NewRegistry(),
	}
	RegisterAction(ctrl.registry, "tap_by_index", "Tap by index", ctrl.TapByIndex)
	return ctrl
}

func (c *Controller) Registry() *Registry {
	return c.registry
}

func (c *Controller) TapByIndex(ctx context.Context, params TapByIndexAction) (*ActionResult, error) {
	return nil, nil
}

func (c *Controller) InputText(ctx context.Context, params InputTextAction) (*ActionResult, error) {
	return nil, nil
}
