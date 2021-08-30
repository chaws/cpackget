/* SPDX-License-Identifier: Apache-2.0 */
/* Copyright Contributors to the cpackget project. */

package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	errs "github.com/open-cmsis-pack/cpackget/cmd/errors"
	log "github.com/sirupsen/logrus"
)

// packNameRegex specifies a regular expression that matches Vendor and Pack names.
// Ref: https://github.com/ARM-software/CMSIS_5/blob/develop/CMSIS/Utilities/PackIndex.xsd
var packNameRegex = regexp.MustCompile(`^[0-9a-zA-Z_\-]+$`)

// versionRegex validates pack version.
// Ref: https://github.com/ARM-software/CMSIS_5/blob/develop/CMSIS/Utilities/PackIndex.xsd
//                                          <major>         . <minor>        . <patch>        - <quality>                                                                                       + <meta info>
var packVersionRegex = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*)?(\+[0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*)?$`)

// IsVendorNameValid checks whether a pack vendor name string matches specified
// regular expression.
func IsPackVendorNameValid(vendorName string) bool {
	return packNameRegex.MatchString(vendorName)
}

// IsPackNameValid checks whether a pack name string matches specified
// regular expression.
func IsPackNameValid(packName string) bool {
	return packNameRegex.MatchString(packName)
}

// IsPackVersion checks whether a pack version string matches specified
// regular expression
func IsPackVersionValid(packVersion string) bool {
	return packVersionRegex.MatchString(packVersion)
}

// PackInfo defines a basic pack information set
type PackInfo struct {
	Location, Vendor, Pack, Version, Extension string
}

// ExtractPackInfo takes in a path to a pack and extracts the needed information.
// It returns an error if any information is wrong
// Valid packPath's are:
// - /path/to/dev/Vendor.Pack.pdsc
// - /path/to/local/Vendor.Pack.Version.pack (or .zip)
// - https://web.com/Vendor.Pack.Version.pack (or .zip)
// If short is true, then prepare it considering that path is in the simpler
// form of Vendor.Pack[.x.y.z], used when removing packs/pdscs.
func ExtractPackInfo(packPath string, short bool) (PackInfo, error) {
	log.Debugf("Extracting pack info from \"%s\"", packPath)

	info := PackInfo{}
	if short {
		_, packName := filepath.Split(packPath)
		details := strings.SplitAfterN(packName, ".", 3)
		if len(details) < 2 {
			return info, errs.ErrBadPackName
		}

		info.Vendor = strings.ReplaceAll(details[0], ".", "")
		info.Pack = strings.ReplaceAll(details[1], ".", "")

		if len(details) == 3 {
			info.Version = details[2]
			if !IsPackVersionValid(info.Version) {
				return info, errs.ErrBadPackNameInvalidVersion
			}
		}

		if !IsPackVendorNameValid(info.Vendor) || !IsPackNameValid(info.Pack) {
			return info, errs.ErrBadPackNameInvalidName
		}

		return info, nil

	}

	validExtensions := map[string]bool{
		".zip":  true,
		".pack": true,
		".pdsc": true,
	}

	location, packName := filepath.Split(packPath)
	info.Extension = filepath.Ext(packName)
	if !validExtensions[info.Extension] {
		return info, errs.ErrBadPackNameInvalidExtension
	}

	isPdsc := info.Extension == ".pdsc"

	details := strings.SplitAfterN(packName, ".", 3)
	if len(details) != 3 {
		return info, errs.ErrBadPackName
	}

	info.Vendor = strings.ReplaceAll(details[0], ".", "")
	info.Pack = strings.ReplaceAll(details[1], ".", "")

	if !isPdsc {
		info.Version = strings.ReplaceAll(details[2], info.Extension, "")
	}

	var err error
	if !IsPackVendorNameValid(info.Vendor) {
		log.Errorf("Pack vendor \"%s\" does not match %s", info.Vendor, packNameRegex)
		err = errs.ErrBadPackNameInvalidVendor
	} else if !IsPackNameValid(info.Pack) {
		log.Errorf("Pack name \"%s\" does not match %s", info.Pack, packNameRegex)
		err = errs.ErrBadPackNameInvalidName
	} else if !isPdsc && !IsPackVersionValid(info.Version) {
		log.Errorf("Pack version \"%s\" does not match %s", info.Version, packVersionRegex)
		err = errs.ErrBadPackNameInvalidVersion
	}

	if err != nil {
		return info, err
	}

	// location can be either a URL or a path to the local
	// file system. If it's the latter, make sure to fill in
	// in case the file is coming from the current directory
	if !(strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") || strings.HasPrefix(location, "file://")) {
		if !filepath.IsAbs(location) {
			absPath, _ := os.Getwd()
			location = filepath.Join(absPath, location)
			location, _ = filepath.Abs(location)
		}

		location = "file://" + location + string(os.PathSeparator)
	}

	info.Location = location
	return info, nil
}
