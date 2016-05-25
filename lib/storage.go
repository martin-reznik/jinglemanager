package lib

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

type data struct {
	Tournament   *Tournament
	Applications []string
	Jingles      []*JingleStorage
}

// Save - will save data to yaml file
func (c *Context) Save() []byte {
	if c.Tournament.Name != "" {
		d := &data{
			Applications: c.Sound.AppNames(),
			Tournament:   c.Tournament,
			Jingles:      c.Jingles.JingleStorageList(),
		}
		out, err := yaml.Marshal(d)
		if err != nil {
			c.Log.Error(err.Error())
		}

		f, err := os.Create(path.Join(c.StorageDir(), "config.yml"))

		if err != nil {
			c.Log.Error(err.Error())
		} else {
			defer f.Close()
			f.Write(out)
		}

		last, _ := os.Create(path.Join(c.AppDir(), "last.tournament"))
		last.Write([]byte(c.Tournament.Name))
		defer last.Close()

		c.Archive()
		return out
	}
	return []byte{}
}

func (c *Context) Archive() {
	target := path.Join(c.AppDir(), fmt.Sprintf("%v.jManager", c.Tournament.Name))
	f, err := os.Create(target)
	if err != nil {
		c.Log.Error(err.Error())
		return
	}
	defer f.Close()
	// Create a buffer to write our archive to.
	buf := bufio.NewWriter(f)

	// Create a new zip archive.
	w := zip.NewWriter(buf)
	defer w.Close()

	// Add some files to the archive.
	var files = []string{
		"config.yml",
	}
	for _, fileName := range c.Songs.FileNames() {
		files = append(files, fmt.Sprintf("media/%v", fileName))
	}
	for _, file := range files {
		zf, err := w.Create(file)
		if err != nil {
			c.Log.Error(err.Error())
		}
		p := path.Join(c.StorageDir(), file)
		c.Log.Info("adding path %v", p)
		data, _ := ioutil.ReadFile(p)
		_, err = zf.Write([]byte(data))
		if err != nil {
			c.Log.Error(err.Error())
		}
	}
}

type addableListI interface {
	AddUniq(string, LogI) (bool, error)
}

// Load - will load data from yaml file
func (c *Context) Load(input []byte) {
	c.cleanup()
	d := &data{}
	yaml.Unmarshal(input, d)

	if d.Tournament == nil {
		return
	}

	c.Tournament = d.Tournament
	c.Tournament.context = c
	c.Tournament.PlanJingles()
	for _, val := range d.Jingles {
		s, err := NewSong(val.File, c)
		if err != nil {
			c.Log.Error(err.Error())
		} else {
			c.Songs.AddUniq(s, c.Log)
			c.Jingles.AddUniq(NewJingle(val.Name, s, val.TimeBeforePoint, val.Point, c), c.Log)
		}
	}
	for _, val := range d.Applications {
		c.Log.Debug("adding application: " + val)
		c.Sound.AddUniq(val, c.Log)
	}
}

// LoadByName - will load data from file
func (c *Context) LoadByName(name string) error {
	if name == "" {
		return errors.New("Nothing to load")
	}
	f, err := os.Open(path.Join(c.AppDir(), name, "config.yml"))
	if err != nil {
		c.Log.Error("File opening error " + name + ": " + err.Error())
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		c.Log.Error("File read error " + name + ": " + err.Error())
		return err
	}
	c.Load(data)
	return nil
}

// SaveSong - will save uploaded song file into tournament directory
func (c *Context) SaveSong(r io.Reader, filename string) (string, error) {
	c.Log.Info(filename)
	targetFile := path.Join(c.MediaDir(), filename)
	writer, err := os.Create(targetFile)

	if err != nil {
		c.Log.Error("File upload failed: " + err.Error())
		return "", err
	}

	defer writer.Close()
	_, err = io.Copy(writer, r)

	if err != nil {
		c.Log.Error("File upload failed: " + err.Error())
		return "", err
	}
	return filename, nil
}

// RemoveSong - will remove song
func (c *Context) RemoveSong(filename string) error {
	filepath := path.Join(c.MediaDir(), filename)
	return os.Remove(filepath)
}

// StorageDir - return path to current tournament directory (and creates path if necessarry)
func (c *Context) StorageDir() string {
	p := path.Join(c.AppDir(), c.Tournament.Name)
	os.MkdirAll(p, 0700)
	return path.Join(p)
}

// MediaDir - return path to current tournament directory (and creates path if necessarry)
func (c *Context) MediaDir() string {
	p := path.Join(c.StorageDir(), "media")
	os.MkdirAll(p, 0700)
	return p
}

// AppDir - return path to application directory
func (c *Context) AppDir() string {
	u, _ := user.Current()
	p := path.Join(u.HomeDir, ".jinglemanager")
	os.MkdirAll(p, 0700)
	return path.Join(p)
}

// LastTournament - return path to application directory
func (c *Context) LastTournament() string {
	f, _ := os.Open(path.Join(c.AppDir(), "last.tournament"))
	t, _ := ioutil.ReadAll(f)
	return string(t)
}
