package static

func (b *Builder) Write(files []*Static) error {
	for _, static := range files {
		// src := static.File.Name()
		// dst := filepath.Join(b.conf.OutputDir, static.Path)
		// b.conf.Log.Debugln("Copying", src, "to", dst)

		if err := b.conf.Write(static.Path, static.File); err != nil {
			return err
		}
	}
	return nil
}
