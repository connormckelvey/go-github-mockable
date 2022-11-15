package generator

type PackageInfo struct {
	Types map[string]*TypeInfo // map[type_name]=>more info
}

func newPackageInfo() *PackageInfo {
	return &PackageInfo{
		Types: make(map[string]*TypeInfo),
	}
}

type TypeInfo struct {
	Name         string
	Doc          string
	FieldDocs    map[string]string   // map[field_name]=>docs
	MethodDocs   map[string]string   //map[method_name]=>docs
	MethodParams map[string][]string // map[method_name]=> []string{}, ie []string{"ctx", "owner", "repo"}
}

func newTypeInfo(name string) *TypeInfo {
	return &TypeInfo{
		Name:         name,
		FieldDocs:    map[string]string{},
		MethodDocs:   map[string]string{},
		MethodParams: map[string][]string{},
	}
}
