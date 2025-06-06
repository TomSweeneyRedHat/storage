package opts

import (
	"fmt"
	"net"
	"slices"
	"strings"
)

// ListOpts holds a list of values and a validation function.
type ListOpts struct {
	values    *[]string
	validator ValidatorFctType
}

// NewListOpts creates a new ListOpts with the specified validator.
func NewListOpts(validator ValidatorFctType) ListOpts {
	var values []string
	return *NewListOptsRef(&values, validator)
}

// NewListOptsRef creates a new ListOpts with the specified values and validator.
func NewListOptsRef(values *[]string, validator ValidatorFctType) *ListOpts {
	return &ListOpts{
		values:    values,
		validator: validator,
	}
}

func (opts *ListOpts) String() string {
	return fmt.Sprintf("%v", *opts.values)
}

// Set validates if needed the input value and adds it to the
// internal slice.
func (opts *ListOpts) Set(value string) error {
	if opts.validator != nil {
		v, err := opts.validator(value)
		if err != nil {
			return err
		}
		value = v
	}
	(*opts.values) = append((*opts.values), value)
	return nil
}

// Delete removes the specified element from the slice.
func (opts *ListOpts) Delete(key string) {
	if i := slices.IndexFunc(*opts.values, func(k string) bool {
		return k == key
	}); i != -1 {
		(*opts.values) = slices.Delete(*opts.values, i, i+1)
	}
}

// GetMap returns the content of values in a map in order to avoid
// duplicates.
func (opts *ListOpts) GetMap() map[string]struct{} {
	ret := make(map[string]struct{})
	for _, k := range *opts.values {
		ret[k] = struct{}{}
	}
	return ret
}

// GetAll returns the values of slice.
func (opts *ListOpts) GetAll() []string {
	return (*opts.values)
}

// GetAllOrEmpty returns the values of the slice
// or an empty slice when there are no values.
func (opts *ListOpts) GetAllOrEmpty() []string {
	v := *opts.values
	if v == nil {
		return make([]string, 0)
	}
	return v
}

// Get checks the existence of the specified key.
func (opts *ListOpts) Get(key string) bool {
	return slices.Contains(*opts.values, key)
}

// Len returns the amount of element in the slice.
func (opts *ListOpts) Len() int {
	return len((*opts.values))
}

// Type returns a string name for this Option type
func (opts *ListOpts) Type() string {
	return "list"
}

// NamedOption is an interface that list and map options
// with names implement.
type NamedOption interface {
	Name() string
}

// NamedListOpts is a ListOpts with a configuration name.
// This struct is useful to keep reference to the assigned
// field name in the internal configuration struct.
type NamedListOpts struct {
	name string
	ListOpts
}

var _ NamedOption = &NamedListOpts{}

// NewNamedListOptsRef creates a reference to a new NamedListOpts struct.
func NewNamedListOptsRef(name string, values *[]string, validator ValidatorFctType) *NamedListOpts {
	return &NamedListOpts{
		name:     name,
		ListOpts: *NewListOptsRef(values, validator),
	}
}

// Name returns the name of the NamedListOpts in the configuration.
func (o *NamedListOpts) Name() string {
	return o.name
}

// MapOpts holds a map of values and a validation function.
type MapOpts struct {
	values    map[string]string
	validator ValidatorFctType
}

// Set validates if needed the input value and add it to the
// internal map, by splitting on '='.
func (opts *MapOpts) Set(value string) error {
	if opts.validator != nil {
		v, err := opts.validator(value)
		if err != nil {
			return err
		}
		value = v
	}
	key, val, _ := strings.Cut(value, "=")
	opts.values[key] = val
	return nil
}

// GetAll returns the values of MapOpts as a map.
func (opts *MapOpts) GetAll() map[string]string {
	return opts.values
}

func (opts *MapOpts) String() string {
	return fmt.Sprintf("%v", opts.values)
}

// Type returns a string name for this Option type
func (opts *MapOpts) Type() string {
	return "map"
}

// NewMapOpts creates a new MapOpts with the specified map of values and a validator.
func NewMapOpts(values map[string]string, validator ValidatorFctType) *MapOpts {
	if values == nil {
		values = make(map[string]string)
	}
	return &MapOpts{
		values:    values,
		validator: validator,
	}
}

// NamedMapOpts is a MapOpts struct with a configuration name.
// This struct is useful to keep reference to the assigned
// field name in the internal configuration struct.
type NamedMapOpts struct {
	name string
	MapOpts
}

var _ NamedOption = &NamedMapOpts{}

// NewNamedMapOpts creates a reference to a new NamedMapOpts struct.
func NewNamedMapOpts(name string, values map[string]string, validator ValidatorFctType) *NamedMapOpts {
	return &NamedMapOpts{
		name:    name,
		MapOpts: *NewMapOpts(values, validator),
	}
}

// Name returns the name of the NamedMapOpts in the configuration.
func (o *NamedMapOpts) Name() string {
	return o.name
}

// ValidatorFctType defines a validator function that returns a validated string and/or an error.
type ValidatorFctType func(val string) (string, error)

// ValidatorFctListType defines a validator function that returns a validated list of string and/or an error
type ValidatorFctListType func(val string) ([]string, error)

// ValidateIPAddress validates an Ip address.
func ValidateIPAddress(val string) (string, error) {
	ip := net.ParseIP(strings.TrimSpace(val))
	if ip != nil {
		return ip.String(), nil
	}
	return "", fmt.Errorf("%s is not an ip address", val)
}

// ValidateLabel validates that the specified string is a valid label, and returns it.
// Labels are in the form on key=value.
func ValidateLabel(val string) (string, error) {
	if strings.Count(val, "=") < 1 {
		return "", fmt.Errorf("bad attribute format: %s", val)
	}
	return val, nil
}

// ValidateSysctl validates a sysctl and returns it.
func ValidateSysctl(val string) (string, error) {
	validSysctlMap := map[string]bool{
		"kernel.msgmax":          true,
		"kernel.msgmnb":          true,
		"kernel.msgmni":          true,
		"kernel.sem":             true,
		"kernel.shmall":          true,
		"kernel.shmmax":          true,
		"kernel.shmmni":          true,
		"kernel.shm_rmid_forced": true,
	}
	validSysctlPrefixes := []string{
		"net.",
		"fs.mqueue.",
	}
	arr := strings.Split(val, "=")
	if len(arr) < 2 {
		return "", fmt.Errorf("sysctl '%s' is not allowed", val)
	}
	if validSysctlMap[arr[0]] {
		return val, nil
	}

	for _, vp := range validSysctlPrefixes {
		if strings.HasPrefix(arr[0], vp) {
			return val, nil
		}
	}
	return "", fmt.Errorf("sysctl '%s' is not allowed", val)
}

// FilterOpt is a flag type for validating filters
type FilterOpt struct {
	filter Args
}

// NewFilterOpt returns a new FilterOpt
func NewFilterOpt() FilterOpt {
	return FilterOpt{filter: NewArgs()}
}

func (o *FilterOpt) String() string {
	repr, err := ToParam(o.filter)
	if err != nil {
		return "invalid filters"
	}
	return repr
}

// Set sets the value of the opt by parsing the command line value
func (o *FilterOpt) Set(value string) error {
	var err error
	o.filter, err = ParseFlag(value, o.filter)
	return err
}

// Type returns the option type
func (o *FilterOpt) Type() string {
	return "filter"
}

// Value returns the value of this option
func (o *FilterOpt) Value() Args {
	return o.filter
}
