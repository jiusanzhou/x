package jsonmerge

// Option ...
type Option func(f *factory)

func (f *factory) With(otps ...Option) Interface {
	for _, o := range otps {
		o(f)
	}
	// TODO: copy a new f
	return f
}

// Overwrite option
func Overwrite(v bool) Option {
	return func(f *factory) {
		f.overwrite = v
	}
}

// OverwriteWithEmptySrc option
func OverwriteWithEmptySrc(v bool) Option {
	return func(f *factory) {
		f.overwriteWithEmptySrc = v
	}
}

// OverwriteSliceWithEmptySrc option
func OverwriteSliceWithEmptySrc(v bool) Option {
	return func(f *factory) {
		f.overwriteSliceWithEmptySrc = v
	}
}

// TypeCheck option
func TypeCheck(v bool) Option {
	return func(f *factory) {
		f.typeCheck = v
	}
}

// AppendSlice option
func AppendSlice(v bool) Option {
	return func(f *factory) {
		f.appendSlice = v
	}
}

// MaxMergeDepth ...
func MaxMergeDepth(v int) Option {
	return func(f *factory) {
		f.maxMergeDepth = v
	}
}
