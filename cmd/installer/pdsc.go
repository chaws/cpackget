/* SPDX-License-Identifier: Apache-2.0 */
/* Copyright Contributors to the cpackget project. */

package installer

import (
	"github.com/open-cmsis-pack/cpackget/cmd/utils"
	"github.com/open-cmsis-pack/cpackget/cmd/xml"
	log "github.com/sirupsen/logrus"
)

// PdscType is the struct that represents the installation of a
// pack via PDSC file
type PdscType struct {
	xml.PdscTag

	// file points to the actual PDSC file
	file *xml.PdscXML

	// path points to a file in the local system, whether or not it's local
	path string
}

// preparePdsc does some sanity validation regarding pdsc name
// and check if it's already installed or not
func preparePdsc(pdscPath string) (*PdscType, error) {
	var err error
	pdsc := &PdscType{
		path: pdscPath,
	}

	info, err := utils.ExtractPackInfo(pdscPath)
	if err != nil {
		return pdsc, err
	}
	pdsc.URL = info.Location
	pdsc.Name = info.Pack
	pdsc.Vendor = info.Vendor
	pdsc.Version = info.Version

	if !Installation.localIsLoaded {
		if err := Installation.LocalPidx.Read(); err != nil {
			return pdsc, err
		}
		Installation.localIsLoaded = true
	}

	return pdsc, err
}

// toPdscTag returns a <pdsc> tag representation of this PDSC file
func (p *PdscType) toPdscTag() (xml.PdscTag, error) {
	tag := p.PdscTag

	if p.file == nil {
		p.file = xml.NewPdscXML(p.path)
		if err := p.file.Read(); err != nil {
			return tag, err
		}
	}

	// uses the Version from the actual file
	tag.Version = p.file.Tag().Version

	return tag, nil
}

// install installs a pack via PDSC file.
// It:
//   - Adds it to the "CMSIS_PACK_ROOT/.Local/local_repository.pidx"
//     using version from the PDSC file
func (p *PdscType) install(installation *PacksInstallationType) error {
	log.Debugf("Installing \"%s\"", p.path)
	tag, err := p.toPdscTag()
	if err != nil {
		return err
	}

	return Installation.LocalPidx.AddPdsc(tag)
}

// uninstall uninstalls a pack via PDSC
// It:
//   - Removes it to the "CMSIS_PACK_ROOT/.Local/local_repository.pidx"
//     If version is ommited, remove all pdsc tags belonging to this pack
func (p *PdscType) uninstall(installation *PacksInstallationType) error {
	log.Debugf("Unistalling \"%s\"", p.path)
	return Installation.LocalPidx.RemovePdsc(p.PdscTag)
}
