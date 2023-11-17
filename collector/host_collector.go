package collector

import (
        "strconv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/tidwall/gjson"
	"github.ibm.com/ZaaS/spectrum-virtualize-exporter/utils"
)

const prefix_host = "spectrum_host_"

var (
	hostVolume *prometheus.Desc
)

func init() {
	registerCollector("lshostvdiskmap", defaultDisabled, NewHostCollector)
	labelnames := []string{"target", "resource", "vdisk_UID", "volume_name", "host_cluster_name"}
	hostVolume = prometheus.NewDesc(prefix_host+"name", "The hosts connected to the storage.", labelnames, nil)
}

//hostCollector collects vdisk metrics
type hostCollector struct {
}

func NewHostCollector() (Collector, error) {
	return &hostCollector{}, nil
}

//Describe describes the metrics
func (*hostCollector) Describe(ch chan<- *prometheus.Desc) {

	ch <- hostVolume

}

//Collect collects metrics from Spectrum Virtualize Restful API
func (c *hostCollector) Collect(sClient utils.SpectrumClient, ch chan<- prometheus.Metric) error {
	log.Debugln("Entering host collector ...")
	reqSystemURL := "https://" + sClient.IpAddress + ":7443/rest/lshostvdiskmap"
	hostResp, err := sClient.CallSpectrumAPI(reqSystemURL)
	if err != nil {
		log.Errorf("Executing lshostvdiskmap cmd failed: %s", err)
	}
	log.Debugln("Response of lshostvdiskmap: ", hostResp)
	// This is a sample output of lshostvdiskmap
	// 	[
	//     {
	//         "id": "0",
	//         "name": "ESXI1",
	//         "SCSI_id": "0",
	//         "vdisk_id": "0",
	//         "vdisk_name": "NeMo_DataStore01",
	//         "vdisk_UID": "6005076380810634F000000000000000",
	//         "IO_group_id": "0",
	//         "IO_group_name": "io_grp0",
	//         "mapping_type": "shared",
	//         "host_cluster_id": "0",
	//         "host_cluster_name": "NEMO_vSphere",
	//         "protocol": "scsi",
	//     }
	// ]

	hostArray := gjson.Parse(hostResp).Array()
	for _, host := range hostArray {
		id_name, err := strconv.ParseFloat(host.Get("id").String(), 64)
		if err != nil {
			log.Errorf("Converting capacity unit failed: %s", err)
				}
		ch <- prometheus.MustNewConstMetric(hostVolume, prometheus.GaugeValue, float64(id_name), sClient.IpAddress, sClient.Hostname, host.Get("vdisk_UID").String(), host.Get("vdisk_name").String(), host.Get("host_cluster_name").String())
	}
	log.Debugln("Leaving host collector.")
	return err

}
