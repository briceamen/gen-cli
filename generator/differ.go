package generator

// DiffServices compares parsed services against the manifest
// and returns only the new methods that need to be generated
func DiffServices(services []Service, manifest *Manifest) map[string][]Method {
	newMethods := make(map[string][]Method)

	for _, svc := range services {
		var methods []Method
		for _, method := range svc.Methods {
			if !manifest.HasMethod(svc.Name, method.Name) {
				methods = append(methods, method)
			}
		}
		if len(methods) > 0 {
			newMethods[svc.Name] = methods
		}
	}

	return newMethods
}

// FindRemovedMethods identifies methods in the manifest that no longer exist in the SDK
func FindRemovedMethods(services []Service, manifest *Manifest) map[string][]string {
	// Build a set of current SDK methods
	sdkMethods := make(map[string]map[string]bool)
	for _, svc := range services {
		sdkMethods[svc.Name] = make(map[string]bool)
		for _, method := range svc.Methods {
			sdkMethods[svc.Name][method.Name] = true
		}
	}

	removed := make(map[string][]string)

	for serviceName, ms := range manifest.Services {
		sdkSvc, ok := sdkMethods[serviceName]
		if !ok {
			// Entire service removed
			for _, method := range ms.Methods {
				removed[serviceName] = append(removed[serviceName], method.Name)
			}
			continue
		}

		for _, method := range ms.Methods {
			if !sdkSvc[method.Name] {
				removed[serviceName] = append(removed[serviceName], method.Name)
			}
		}
	}

	return removed
}
