package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/storage/eventdb"
	"github.com/eleme/banshee/storage/metricdb"
)

const (
	backupDirExt = "_backup"
	oldDirExt    = "_old"
)

// Migrate data when configuration has been changed.
func Migrate(fileName string, opts *Options) error {
	oldOpts := &Options{}
	lockFilePath := path.Join(fileName, optionlockFileName)
	b, err := ioutil.ReadFile(lockFilePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, oldOpts)
	if err != nil {
		return err
	}
	if opts.Period == oldOpts.Period {
		return nil
	}
	newFileName := fileName + backupDirExt
	_, err = os.Stat(newFileName)
	if err == nil {
		err := os.RemoveAll(newFileName)
		if err != nil {
			return err
		}
	}
	err = os.Mkdir(newFileName, filemode)
	if err != nil {
		return err
	}
	defer os.RemoveAll(newFileName)
	err = migrateAdminDB(newFileName, fileName)
	if err != nil {
		return err
	}
	err = migrateIndexDB(newFileName, fileName)
	if err != nil {
		return err
	}
	err = migrateMetricDB(newFileName, opts, fileName, oldOpts)
	if err != nil {
		return err
	}
	err = migrateEventDB(newFileName, opts, fileName, oldOpts)
	if err != nil {
		return err
	}
	b, _ = yaml.Marshal(opts)
	err = ioutil.WriteFile(path.Join(newFileName, optionlockFileName), b, 0644)
	if err != nil {
		return err
	}
	err = os.Rename(fileName, fileName+oldDirExt)
	if err != nil {
		return err
	}
	err = os.Rename(newFileName, fileName)
	if err != nil {
		return err
	}
	return nil
}

func migrateAdminDB(fileName string, oldFileName string) error {
	src := path.Join(oldFileName, admindbFileName)
	dst := path.Join(fileName, admindbFileName)
	return copyFile(src, dst)
}

func migrateIndexDB(fileName string, oldFileName string) error {
	src := path.Join(oldFileName, indexdbFileName)
	dst := path.Join(fileName, indexdbFileName)
	return copyDir(src, dst)
}

func migrateMetricDB(fileName string, opts *Options, oldFileName string, oldOpts *Options) error {
	var option, oldOption *metricdb.Options
	if opts != nil {
		option = &metricdb.Options{
			Period:       opts.Period,
			Expiration:   opts.Expiration,
			FilterOffset: opts.FilterOffset,
		}
	}
	metric, err := metricdb.Open(path.Join(fileName, metricdbFileName), option)
	if err != nil {
		return err
	}
	if oldOpts != nil {
		oldOption = &metricdb.Options{
			Period:       oldOpts.Period,
			Expiration:   oldOpts.Expiration,
			FilterOffset: oldOpts.FilterOffset,
		}
	}
	oldMetric, err := metricdb.Open(path.Join(oldFileName, metricdbFileName), oldOption)
	if err != nil {
		return err
	}
	err = oldMetric.Scan(func(m *models.Metric) error {
		err := metric.Put(m)
		if err == nil || err == metricdb.ErrNoStorage {
			return nil
		}
		return err
	})
	return nil
}

func migrateEventDB(fileName string, opts *Options, oldFileName string, oldOpts *Options) error {
	var option, oldOption *eventdb.Options
	if opts != nil {
		option = &eventdb.Options{
			Period:     opts.Period,
			Expiration: opts.Expiration,
		}
	}
	event, err := eventdb.Open(path.Join(fileName, eventdbFileName), option)
	if err != nil {
		return err
	}
	if oldOpts != nil {
		oldOption = &eventdb.Options{
			Period:     oldOpts.Period,
			Expiration: oldOpts.Expiration,
		}
	}
	oldEvent, err := eventdb.Open(path.Join(oldFileName, eventdbFileName), oldOption)
	if err != nil {
		return err
	}
	err = oldEvent.Scan(func(ew eventdb.EventWrapper) error {
		err := event.Put(&ew)
		if err == nil || err == eventdb.ErrNoStorage {
			return nil
		}
		return err
	})
	return nil
}

func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}
