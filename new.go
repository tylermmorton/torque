package torque

import "github.com/pkg/errors"

func New[T ViewModel](ctl Controller) (Handler, error) {
	var (
		// vm is the zero value of the generic constraint that
		// can be used in type assertions
		vm  ViewModel = new(T)
		h             = createHandlerImpl[T](ctl)
		err error
	)

	err = assertImplementations(h, ctl, vm)
	if err != nil {
		return nil, errors.Wrap(err, "failed to assert Controller interface")
	}

	for _, plugin := range h.plugins {
		err = plugin.Setup(h)(ctl, vm)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to install Plugin %T", plugin)
		}
	}

	return h, nil
}

func MustNew[T ViewModel](ctl Controller) Handler {
	h, err := New[T](ctl)
	if err != nil {
		panic(err)
	}
	return h
}
