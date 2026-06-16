package server

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func (s *Server) watchDir(watcher *fsnotify.Watcher, path string) error {
	return filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(p, ".") || strings.HasPrefix(p, "_") {
				return fs.SkipDir
			}
			return watcher.Add(p)
		}
		return nil
	})
}

func (s *Server) watch(paths []string, fn func(string, fs.FileInfo)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	for _, path := range paths {
		if _, err := os.Stat(path); err != nil && !os.IsExist(err) {
			continue
		}
		if err := s.watchDir(watcher, path); err != nil {
			s.ctx.Logger.Errorf("Error watching %s: %v", path, err)
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			info, err := os.Stat(event.Name)
			if err != nil {
				continue
			}
			if info.IsDir() {
				continue
			}

			if event.Has(fsnotify.Write) {
				s.ctx.Logger.Infoln("The", event.Name, "has been modified. Rebuilding...")

				fn(event.Name, info)
			} else if event.Has(fsnotify.Create) {
				s.ctx.Logger.Infoln("The", event.Name, "has been created. Rebuilding...")

				fn(event.Name, info)
			} else if event.Has(fsnotify.Remove) {
				s.ctx.Logger.Infoln("The", event.Name, "has been removed. Rebuilding...")

				fn(event.Name, info)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("监听错误: %v", err)
		}
	}
}

func (s *Server) watchFiles() {
	staticDirs := []string{"static"}
	templatesDirs := []string{"templates"}
	if theme := s.ctx.GetTheme(); theme != "" {
		staticDirs = append(staticDirs, filepath.Join("themes", theme, "static"))
		templatesDirs = append(templatesDirs, filepath.Join("themes", theme, "templates"))
	}
	contentDir := s.ctx.GetContentDir()

	watchDirs := slices.Concat([]string{contentDir}, staticDirs, templatesDirs)
	if err := s.watch(watchDirs, func(file string, info fs.FileInfo) {
		if strings.HasPrefix(file, contentDir+"/") {
			if err := s.reloadContent(file, info); err != nil {
				s.ctx.Logger.Errorf("Reload content err: %s", err.Error())
			}
			return
		}
		for _, staticDir := range staticDirs {
			if strings.HasPrefix(file, staticDir+"/") {
				if err := s.reloadStatic(staticDir, file, info); err != nil {
					s.ctx.Logger.Errorf("Reload static err: %s", err.Error())
				}
				return
			}
		}
		for _, templateDir := range templatesDirs {
			if strings.HasPrefix(file, templateDir+"/") {
				if err := s.reloadTemplate(file, info); err != nil {
					s.ctx.Logger.Errorf("Reload templates err: %s", err.Error())
				}
				return
			}
		}
	}); err != nil {
		s.ctx.Logger.Errorf("Watch files err: %s", err.Error())
	}
}

func (s *Server) reloadContent(file string, info fs.FileInfo) error {
	contentPath, err := filepath.Rel(s.ctx.GetContentDir(), file)
	if err != nil {
		return err
	}
	if s.site.IsIgnoredContent(filepath.ToSlash(contentPath), info.IsDir()) {
		return nil
	}

	s.fs.Reset()

	if err := s.site.BuildContent(context.TODO(), s.fs); err != nil {
		return err
	}
	if s.livereload != nil {
		s.livereload.Notify("*.html")
	}
	return nil
}

func (s *Server) reloadStatic(baseDir string, file string, info fs.FileInfo) error {
	srcPath, err := filepath.Rel(baseDir, file)
	if err != nil {
		return err
	}
	if s.site.IsIgnoredStatic(filepath.ToSlash(srcPath), info.IsDir()) {
		return nil
	}

	staticFS, err := s.ctx.GetFS("static", true)
	if err != nil {
		return err
	}

	src, err := staticFS.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dstPath := srcPath
	if !strings.HasPrefix(dstPath, "/") {
		dstPath = "/" + dstPath
	}
	if err := s.fs.WriteFile(context.TODO(), dstPath, src); err != nil {
		return err
	}
	if s.livereload != nil {
		s.livereload.Notify(dstPath)
	}
	return nil
}

func (s *Server) reloadTemplate(file string, info fs.FileInfo) error {
	s.fs.Reset()

	if err := s.site.Build(context.TODO(), s.fs); err != nil {
		return err
	}
	if s.livereload != nil {
		s.livereload.Notify("*.html")
	}
	return nil
}
