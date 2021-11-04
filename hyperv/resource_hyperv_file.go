package hyperv

import (
	"context"
	"fmt"
	"log"

	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/taliesins/terraform-provider-hyperv/api"
)

func resourceHyperVFile() *schema.Resource {
	return &schema.Resource{
		Create: resourceHyperVFileCreate,
		Read:   resourceHyperVFileRead,
		Update: resourceHyperVFileUpdate,
		Delete: resourceHyperVFileDelete,

		Schema: map[string]*schema.Schema{
			"path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"source": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"exists": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},

		CustomizeDiff: customizeDiffForFile,
	}
}

func customizeDiffForFile(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
	path := diff.Get("path").(string)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			diff.SetNewComputed("exists")
			return nil
		} else {
			// other error
			return err
		}
	}

	return nil
}

func resourceHyperVFileCreate(d *schema.ResourceData, meta interface{}) (err error) {

	log.Printf("[INFO][hyperv][create] creating hyperv file: %#v", d)
	c := meta.(*api.HypervClient)

	path := ""

	if v, ok := d.GetOk("path"); ok {
		path = v.(string)
	} else {
		return fmt.Errorf("[ERROR][hyperv][create] path argument is required")
	}

	source := (d.Get("source")).(string)

	err = c.CreateOrUpdateFile(path, source)

	if err != nil {
		return err
	}
	d.SetId(path)

	log.Printf("[INFO][hyperv][create] created hyperv file: %#v", d)

	return resourceHyperVFileRead(d, meta)
}

func resourceHyperVFileRead(d *schema.ResourceData, meta interface{}) (err error) {
	log.Printf("[INFO][hyperv][read] reading hyperv vhd: %#v", d)
	c := meta.(*api.HypervClient)

	path := ""

	if v, ok := d.GetOk("path"); ok {
		path = v.(string)
	} else {
		return fmt.Errorf("[ERROR][hyperv][read] path argument is required")
	}

	vhd, err := c.GetVhd(path)
	if err != nil {
		return err
	}

	d.SetId(path)
	d.Set("path", vhd.Path)

	if vhd.Path != "" {
		log.Printf("[INFO][hyperv][read] unable to retrieved vhd: %+v", path)
		d.Set("exists", false)
	} else {
		log.Printf("[INFO][hyperv][read] retrieved vhd: %+v", path)
		d.Set("exists", true)
	}

	log.Printf("[INFO][hyperv][read] read hyperv vhd: %#v", d)

	return nil
}

func resourceHyperVFileUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	log.Printf("[INFO][hyperv][update] updating hyperv vhd: %#v", d)
	c := meta.(*api.HypervClient)

	path := ""

	if v, ok := d.GetOk("path"); ok {
		path = v.(string)
	} else {
		return fmt.Errorf("[ERROR][hyperv][update] path argument is required")
	}

	source := (d.Get("source")).(string)
	sourceVm := (d.Get("source_vm")).(string)
	sourceDisk := (d.Get("source_disk")).(int)
	vhdType := api.ToVhdType((d.Get("vhd_type")).(string))
	parentPath := (d.Get("parent_path")).(string)
	size := uint64((d.Get("size")).(int))
	blockSize := uint32((d.Get("block_size")).(int))
	logicalSectorSize := uint32((d.Get("logical_sector_size")).(int))
	physicalSectorSize := uint32((d.Get("physical_sector_size")).(int))

	exists := (d.Get("exists")).(bool)

	if !exists || d.HasChange("path") || d.HasChange("source") || d.HasChange("source_vm") || d.HasChange("source_disk") || d.HasChange("parent_path") {
		//delete it as its changed
		err = c.CreateOrUpdateVhd(path, source, sourceVm, sourceDisk, vhdType, parentPath, size, blockSize, logicalSectorSize, physicalSectorSize)

		if err != nil {
			return err
		}
	}

	if size > 0 && parentPath == "" {
		if !exists || d.HasChange("size") {
			//Update vhd size
			err = c.ResizeVhd(path, size)

			if err != nil {
				return err
			}
		}
	}

	log.Printf("[INFO][hyperv][update] updated hyperv vhd: %#v", d)

	return resourceHyperVFileRead(d, meta)
}

func resourceHyperVFileDelete(d *schema.ResourceData, meta interface{}) (err error) {
	log.Printf("[INFO][hyperv][delete] deleting hyperv vhd: %#v", d)

	c := meta.(*api.HypervClient)

	path := ""

	if v, ok := d.GetOk("path"); ok {
		path = v.(string)
	} else {
		return fmt.Errorf("[ERROR][hyperv][delete] path argument is required")
	}

	err = c.DeleteVhd(path)

	if err != nil {
		return err
	}

	log.Printf("[INFO][hyperv][delete] deleted hyperv vhd: %#v", d)
	return nil
}
