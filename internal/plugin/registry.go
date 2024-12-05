package plugin

type Registry struct {
	sources map[string]func() SourcePlugin
	targets map[string]func() TargetPlugin
}

func NewRegistry() *Registry {
	return &Registry{
		sources: make(map[string]func() SourcePlugin),
		targets: make(map[string]func() TargetPlugin),
	}
}

func (r *Registry) RegisterSource(name string, factory func() SourcePlugin) {
	r.sources[name] = factory
}

func (r *Registry) RegisterTarget(name string, factory func() TargetPlugin) {
	r.targets[name] = factory
}

func (r *Registry) GetSource(name string) (SourcePlugin, bool) {
	factory, exists := r.sources[name]
	if !exists {
		return nil, false
	}
	return factory(), true
}

func (r *Registry) GetTarget(name string) (TargetPlugin, bool) {
	factory, exists := r.targets[name]
	if !exists {
		return nil, false
	}
	return factory(), true
}
