package databinding

// BindingAction selects the behaviour of WritableValue.UpdateFrom and
// Property.Bind at the moment the binding is created. Sealed: only
// UpdateOnBind and NoActionOnBind are valid values.
type BindingAction interface {
	isBindingAction()
}

type updateOnBind struct{}

func (updateOnBind) isBindingAction() {}

type noActionOnBind struct{}

func (noActionOnBind) isBindingAction() {}

// UpdateOnBind requests an immediate refresh of the target value when
// the binding is established.
var UpdateOnBind BindingAction = updateOnBind{}

// NoActionOnBind requests no refresh; the target only updates when
// the source next changes.
var NoActionOnBind BindingAction = noActionOnBind{}
