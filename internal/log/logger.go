package log

// Init creates new logger from given configuration and set it as the global [Logger].
func Init(c Config) error {
	lvl, err := c.level()
	if err != nil {
		return err
	}

	w, err := c.writer()
	if err != nil {
		return err
	}

	Logger = Logger.Output(w).Level(lvl)
	return nil
}
