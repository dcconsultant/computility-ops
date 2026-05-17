package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"computility-ops/backend/internal/repository"
)

type ReconfigTargetConfig struct {
	Mode                    string  `json:"mode,omitempty"`
	ConfigType              string  `json:"config_type,omitempty"`
	PerfBaseline            float64 `json:"perf_baseline"`
	MemoryDatarateBaseline  float64 `json:"memory_datarate_baseline"`
	MemoryCapacityBaseline  float64 `json:"memory_capacity_baseline"`
	StorageCapacityBaseline float64 `json:"storage_capacity_baseline"`
	MemoryCPURatio          float64 `json:"memory_cpu_ratio,omitempty"`
}

type ReconfigScopeConfig struct {
	PSAList     []string `json:"psa_list"`
	ConfigTypes []string `json:"config_types"`
	SNs         []string `json:"sns"`
}

type ReconfigPlanCalculateRequest struct {
	Target         ReconfigTargetConfig `json:"target"`
	Scope          ReconfigScopeConfig  `json:"scope"`
	GoalValueScore float64              `json:"goal_value_score,omitempty"`
}

type ReconfigCandidateRow struct {
	SN               string  `json:"sn"`
	ConfigType       string  `json:"config_type"`
	Rack             string  `json:"rack"`
	Datacenter       string  `json:"datacenter"`
	MemoryDatarate   float64 `json:"memory_datarate"`
	PerfScore        float64 `json:"perf_score"`
	MemoryCapacityGB float64 `json:"memory_capacity_gb"`
	StorageCapacityTB float64 `json:"storage_capacity_tb"`
	MemoryGapGB      float64 `json:"memory_gap_gb"`
	StorageGapTB     float64 `json:"storage_gap_tb"`
	Status           string  `json:"status"`
	ValueScoreWarn   bool    `json:"value_score_warn,omitempty"`
}

type ReconfigActionRow struct {
	TargetSN    string `json:"target_sn"`
	GapType     string `json:"gap_type"`
	GapQty      string `json:"gap_qty"`
	Source      string `json:"source"`
	PartDetails string `json:"part_details"`
	CrossIDC    string `json:"cross_idc"`
	Action      string `json:"action"`
	RuleHit     string `json:"rule_hit"`
}

type ReconfigPlanProgress struct {
	Stage        string  `json:"stage"`
	Percent      float64 `json:"percent"`
	Message      string  `json:"message,omitempty"`
	DonePackages int     `json:"done_packages"`
	TotalPackages int    `json:"total_packages"`
	DoneServers  int     `json:"done_servers"`
	TotalServers int     `json:"total_servers"`
	DoneCores    float64 `json:"done_cores"`
	TotalCores   float64 `json:"total_cores"`
}

type ReconfigPlanCalculateResponse struct {
	TargetResolved ReconfigTargetConfig   `json:"target_resolved"`
	Candidates     []ReconfigCandidateRow `json:"candidates"`
	Actions        []ReconfigActionRow    `json:"actions"`
	Summary        map[string]any         `json:"summary"`
}

type ReconfigService struct {
	metaRepo    repository.MetaRepo
	datasetRepo repository.DatasetRepo
}

func NewReconfigService(metaRepo repository.MetaRepo, datasetRepo repository.DatasetRepo) *ReconfigService {
	return &ReconfigService{metaRepo: metaRepo, datasetRepo: datasetRepo}
}

func (s *ReconfigService) CalculatePlan(ctx context.Context, req ReconfigPlanCalculateRequest) (ReconfigPlanCalculateResponse, error) {
	return s.CalculatePlanWithProgress(ctx, req, nil)
}

func (s *ReconfigService) CalculatePlanWithProgress(ctx context.Context, req ReconfigPlanCalculateRequest, onProgress func(ReconfigPlanProgress)) (ReconfigPlanCalculateResponse, error) {
	servers, racks, memories, disks, configs, err := s.loadMetaRecords(ctx)
	if err != nil {
		return ReconfigPlanCalculateResponse{}, err
	}
	perfRows, err := s.datasetRepo.ListValueScorePerformanceParams(ctx)
	if err != nil {
		return ReconfigPlanCalculateResponse{}, err
	}
	perfByConfig := map[string]float64{}
	for _, x := range perfRows {
		k := strings.TrimSpace(x.ConfigType)
		if k == "" {
			continue
		}
		perfByConfig[k] = x.PerformanceScore
	}

	rackIDC := map[string]string{}
	for _, r := range racks {
		rackNo := strings.TrimSpace(pick(r, "sn", "rack", "机柜编号"))
		if rackNo == "" {
			continue
		}
		rackIDC[rackNo] = strings.TrimSpace(pick(r, "datacenter", "idc", "机房"))
	}

	target := req.Target
	if strings.TrimSpace(target.ConfigType) != "" {
		cfgType := strings.TrimSpace(target.ConfigType)
		if target.PerfBaseline <= 0 {
			target.PerfBaseline = perfByConfig[cfgType]
		}
		if target.MemoryCapacityBaseline <= 0 || target.StorageCapacityBaseline <= 0 {
			for _, c := range configs {
				if strings.TrimSpace(pick(c, "config_type", "配置类型")) == cfgType {
					if target.MemoryCapacityBaseline <= 0 {
						target.MemoryCapacityBaseline = pickNum(c, "capacity_memory_gb", "内存容量(GB)")
					}
					if target.StorageCapacityBaseline <= 0 {
						target.StorageCapacityBaseline = pickNum(c, "capacity_storage_tb", "存储容量(TB)")
					}
					break
				}
			}
		}
		if target.MemoryDatarateBaseline <= 0 {
			var sampleSN string
			for _, sv := range servers {
				if strings.TrimSpace(pick(sv, "config_type", "配置类型")) == cfgType {
					sampleSN = strings.TrimSpace(pick(sv, "sn", "SN"))
					break
				}
			}
			if sampleSN != "" {
				for _, mem := range memories {
					if strings.TrimSpace(pick(mem, "sn_server", "服务器SN")) == sampleSN {
						target.MemoryDatarateBaseline = pickNum(mem, "datarate", "数据传输率(TM/s)", "数据传输率(MT/s)")
						break
					}
				}
			}
		}
	}

	snFilter := map[string]struct{}{}
	for _, sn := range req.Scope.SNs {
		s := strings.TrimSpace(sn)
		if s != "" {
			snFilter[s] = struct{}{}
		}
	}
	cfgFilter := map[string]struct{}{}
	for _, c := range req.Scope.ConfigTypes {
		k := strings.TrimSpace(c)
		if k != "" {
			cfgFilter[k] = struct{}{}
		}
	}

	filteredServers := make([]map[string]any, 0)
	for _, sv := range servers {
		psa := strings.TrimSpace(pick(sv, "psa", "PSA"))
		belong := strings.TrimSpace(pick(sv, "belong", "归属"))
		cfgType := strings.TrimSpace(pick(sv, "config_type", "配置类型"))
		sn := strings.TrimSpace(pick(sv, "sn", "SN"))

		if belong != "" && belong != "ai_data_engineering" {
			continue
		}
		if len(req.Scope.PSAList) > 0 {
			hit := false
			for _, p := range req.Scope.PSAList {
				if strings.Contains(psa, strings.TrimSpace(p)) {
					hit = true
					break
				}
			}
			if !hit {
				continue
			}
		}
		if len(cfgFilter) > 0 {
			if _, ok := cfgFilter[cfgType]; !ok {
				continue
			}
		}
		if len(snFilter) > 0 {
			if _, ok := snFilter[sn]; !ok {
				continue
			}
		}
		filteredServers = append(filteredServers, sv)
	}

	coreByConfig := map[string]float64{}
	for _, c := range configs {
		cfg := strings.TrimSpace(pick(c, "config_type", "配置类型"))
		if cfg == "" {
			continue
		}
		coreByConfig[cfg] = pickNum(c, "logical_cores", "逻辑核")
	}
	packageTotalByCore := map[string]int{}
	totalCores := 0.0
	for _, sv := range filteredServers {
		cfg := strings.TrimSpace(pick(sv, "config_type", "配置类型"))
		core := coreByConfig[cfg]
		key := fmt.Sprintf("%.0f", core)
		packageTotalByCore[key]++
		totalCores += core
	}
	totalPackages := len(packageTotalByCore)
	if onProgress != nil {
		onProgress(ReconfigPlanProgress{Stage: "candidate", Percent: 5, Message: "开始生成候选", TotalPackages: totalPackages, TotalServers: len(filteredServers), TotalCores: totalCores})
	}
	packageDoneByCore := map[string]int{}
	donePackages := 0
	doneServers := 0
	doneCores := 0.0

	candidates := make([]ReconfigCandidateRow, 0, len(filteredServers))
	for _, sv := range filteredServers {
		sn := strings.TrimSpace(pick(sv, "sn", "SN"))
		cfgType := strings.TrimSpace(pick(sv, "config_type", "配置类型"))
		rack := strings.TrimSpace(pick(sv, "rack", "机柜"))
		idc := strings.TrimSpace(pick(sv, "idc", "机房"))
		if idc == "" {
			idc = rackIDC[rack]
		}

		hostMems := filterBySN(memories, sn)
		hostDisks := filterBySN(disks, sn)
		memRate := effectiveMemoryDatarate(hostMems)
		perf := perfByConfig[cfgType]
		memCap := sumNum(hostMems, "capacity", "容量")
		storageTB := sumNum(hostDisks, "capacity", "容量") / 1024.0

		memBaseline := target.MemoryCapacityBaseline
		storageBaseline := target.StorageCapacityBaseline
		if strings.EqualFold(strings.TrimSpace(target.Mode), "maximize") {
			ratio := target.MemoryCPURatio
			if ratio <= 0 {
				ratio = 6
			}
			logicalCores := 0.0
			for _, c := range configs {
				if strings.TrimSpace(pick(c, "config_type", "配置类型")) == cfgType {
					logicalCores = pickNum(c, "logical_cores", "逻辑核")
					break
				}
			}
			if logicalCores > 0 {
				memBaseline = logicalCores * ratio
			}
			storageBaseline = 0
		}
		memGap := math.Max(0, memBaseline-memCap)
		storageGap := math.Max(0, storageBaseline-storageTB)

		status := "候选"
		if !strings.EqualFold(strings.TrimSpace(target.Mode), "maximize") {
			if memRate < target.MemoryDatarateBaseline {
				status = "内存带宽不足"
			} else if perf < target.PerfBaseline {
				status = "性能不足"
			}
		}

		warn := req.GoalValueScore > 0 && perf < req.GoalValueScore
		doneServers++
		doneCores += coreByConfig[cfgType]
		coreKey := fmt.Sprintf("%.0f", coreByConfig[cfgType])
		if packageTotalByCore[coreKey] > 0 {
			packageDoneByCore[coreKey]++
			if packageDoneByCore[coreKey] == packageTotalByCore[coreKey] {
				donePackages++
			}
		}
		if onProgress != nil && (doneServers%20 == 0 || doneServers == len(filteredServers)) {
			percent := 5.0
			if len(filteredServers) > 0 {
				percent = 5 + float64(doneServers)*55/float64(len(filteredServers))
			}
			onProgress(ReconfigPlanProgress{Stage: "candidate", Percent: percent, Message: "候选生成中", DonePackages: donePackages, TotalPackages: totalPackages, DoneServers: doneServers, TotalServers: len(filteredServers), DoneCores: round2(doneCores), TotalCores: round2(totalCores)})
		}
		candidates = append(candidates, ReconfigCandidateRow{
			SN:                sn,
			ConfigType:        cfgType,
			Rack:              rack,
			Datacenter:        idc,
			MemoryDatarate:    round2(memRate),
			PerfScore:         round2(perf),
			MemoryCapacityGB:  round2(memCap),
			StorageCapacityTB: round2(storageTB),
			MemoryGapGB:       round2(memGap),
			StorageGapTB:      round2(storageGap),
			Status:            status,
			ValueScoreWarn:    warn,
		})
	}

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].SN < candidates[j].SN })
	if onProgress != nil {
		onProgress(ReconfigPlanProgress{Stage: "action", Percent: 70, Message: "生成执行清单中", DonePackages: donePackages, TotalPackages: totalPackages, DoneServers: doneServers, TotalServers: len(filteredServers), DoneCores: round2(doneCores), TotalCores: round2(totalCores)})
	}
	actions := s.buildActions(candidates, filteredServers, memories, disks, rackIDC)
	if onProgress != nil {
		onProgress(ReconfigPlanProgress{Stage: "done", Percent: 100, Message: "计算完成", DonePackages: donePackages, TotalPackages: totalPackages, DoneServers: doneServers, TotalServers: len(filteredServers), DoneCores: round2(doneCores), TotalCores: round2(totalCores)})
	}

	return ReconfigPlanCalculateResponse{
		TargetResolved: target,
		Candidates:     candidates,
		Actions:        actions,
		Summary: map[string]any{
			"scope_server_count": len(filteredServers),
			"candidate_count":    len(candidates),
			"action_count":       len(actions),
		},
	}, nil
}

func (s *ReconfigService) loadMetaRecords(ctx context.Context) (servers, racks, memories, disks, configs []map[string]any, err error) {
	modelCodes := []string{"server", "rack", "memory", "disk", "config_type"}
	data := map[string][]map[string]any{}
	for _, code := range modelCodes {
		m, e := s.metaRepo.GetModelByCode(ctx, code)
		if e != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("读取模型失败(%s): %w", code, e)
		}
		records, e := s.metaRepo.ListRecords(ctx, m.ID)
		if e != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("读取模型记录失败(%s): %w", code, e)
		}
		rows := make([]map[string]any, 0, len(records))
		for _, r := range records {
			rows = append(rows, r.Data)
		}
		data[code] = rows
	}
	return data["server"], data["rack"], data["memory"], data["disk"], data["config_type"], nil
}

func (s *ReconfigService) buildActions(candidates []ReconfigCandidateRow, filteredServers []map[string]any, memories, disks []map[string]any, rackIDC map[string]string) []ReconfigActionRow {
	serverRack := map[string]string{}
	serverIDC := map[string]string{}
	for _, sv := range filteredServers {
		sn := strings.TrimSpace(pick(sv, "sn", "SN"))
		rack := strings.TrimSpace(pick(sv, "rack", "机柜"))
		idc := strings.TrimSpace(pick(sv, "idc", "机房"))
		if idc == "" {
			idc = strings.TrimSpace(rackIDC[rack])
		}
		if sn != "" {
			serverRack[sn] = rack
			serverIDC[sn] = idc
		}
	}

	candidateStatus := map[string]string{}
	for _, c := range candidates {
		candidateStatus[c.SN] = c.Status
	}

	memoryInventory := make([]map[string]any, 0)
	for _, m := range memories {
		if strings.TrimSpace(pick(m, "sn_server", "服务器SN")) != "" {
			continue
		}
		rack := strings.TrimSpace(pick(m, "rack", "机柜"))
		if strings.Contains(strings.ToUpper(rack), "SPR") {
			memoryInventory = append(memoryInventory, m)
		}
	}
	diskInventory := make([]map[string]any, 0)
	for _, d := range disks {
		if strings.TrimSpace(pick(d, "sn_server", "服务器SN")) != "" {
			continue
		}
		rack := strings.TrimSpace(pick(d, "rack", "机柜"))
		if strings.Contains(strings.ToUpper(rack), "SPR") {
			diskInventory = append(diskInventory, d)
		}
	}

	actions := make([]ReconfigActionRow, 0)
	for _, c := range candidates {
		if c.Status != "候选" || (c.MemoryGapGB <= 0 && c.StorageGapTB <= 0) {
			continue
		}
		hostMems := filterBySN(memories, c.SN)
		hostDisks := filterBySN(disks, c.SN)
		memSpec := map[string]any{}
		diskSpec := map[string]any{}
		if len(hostMems) > 0 {
			memSpec = hostMems[0]
		}
		if len(hostDisks) > 0 {
			diskSpec = hostDisks[0]
		}
		memUnit := math.Max(1, pickNum(memSpec, "capacity", "容量"))
		diskUnitTB := math.Max(0.1, pickNum(diskSpec, "capacity", "容量")/1024)
		needMemCount := int(math.Ceil(c.MemoryGapGB / memUnit))
		needDiskCount := int(math.Ceil(c.StorageGapTB / diskUnitTB))

		if needMemCount > 0 {
			matched := make([]map[string]any, 0)
			for _, m := range memoryInventory {
				if strings.TrimSpace(pick(m, "brand", "厂商")) != strings.TrimSpace(pick(memSpec, "brand", "厂商")) {
					continue
				}
				if strings.TrimSpace(pick(m, "model", "型号")) != strings.TrimSpace(pick(memSpec, "model", "型号")) {
					continue
				}
				if pickNum(m, "capacity", "容量") != pickNum(memSpec, "capacity", "容量") {
					continue
				}
				if pickNum(m, "datarate", "数据传输率(TM/s)", "数据传输率(MT/s)") != pickNum(memSpec, "datarate", "数据传输率(TM/s)", "数据传输率(MT/s)") {
					continue
				}
				matched = append(matched, m)
				if len(matched) >= needMemCount {
					break
				}
			}

			if len(matched) >= needMemCount {
				srcRack := strings.TrimSpace(pick(matched[0], "rack", "机柜"))
				srcIDC := strings.TrimSpace(rackIDC[srcRack])
				sns := make([]string, 0, len(matched))
				for _, x := range matched {
					sns = append(sns, strings.TrimSpace(pick(x, "sn", "序列号")))
				}
				actions = append(actions, ReconfigActionRow{
					TargetSN:    c.SN,
					GapType:     "内存",
					GapQty:      fmt.Sprintf("%d 条", needMemCount),
					Source:      fmt.Sprintf("库房（机房:%s 机柜:%s）", srcIDC, srcRack),
					PartDetails: strings.Join(sns, ", "),
					CrossIDC:    ternary(srcIDC != "" && c.Datacenter != "" && srcIDC != c.Datacenter, "是", "否"),
					Action:      "改配",
					RuleHit:     "内存规格一致（厂商/型号/容量/速率）+ 优先库房资源",
				})
			} else {
				donor := s.pickMemoryDonor(memories, memSpec, c.SN, needMemCount, candidateStatus, serverRack, serverIDC)
				crossIDC := "否"
				if donor.anyIDC != "" && c.Datacenter != "" && donor.anyIDC != c.Datacenter {
					crossIDC = "是"
				}
				partDetails := donor.details
				if donor.missingCount > 0 {
					partDetails = fmt.Sprintf("%s；仍缺 %d 条", donor.details, donor.missingCount)
				}
				actions = append(actions, ReconfigActionRow{
					TargetSN:    c.SN,
					GapType:     "内存",
					GapQty:      fmt.Sprintf("%d 条", needMemCount),
					Source:      donor.source,
					PartDetails: partDetails,
					CrossIDC:    crossIDC,
					Action:      "调拨",
					RuleHit:     "库房不足后按优先级从主机侧调拨（内存不从内存带宽不足主机取件）",
				})
			}
		}

		if needDiskCount > 0 {
			matched := make([]map[string]any, 0)
			for _, d := range diskInventory {
				if strings.TrimSpace(pick(d, "brand", "厂商")) != strings.TrimSpace(pick(diskSpec, "brand", "厂商")) {
					continue
				}
				if strings.TrimSpace(pick(d, "model", "型号")) != strings.TrimSpace(pick(diskSpec, "model", "型号")) {
					continue
				}
				if pickNum(d, "capacity", "容量") != pickNum(diskSpec, "capacity", "容量") {
					continue
				}
				if strings.TrimSpace(pick(d, "application_category", "应用分类")) != strings.TrimSpace(pick(diskSpec, "application_category", "应用分类")) {
					continue
				}
				matched = append(matched, d)
				if len(matched) >= needDiskCount {
					break
				}
			}

			if len(matched) >= needDiskCount {
				srcRack := strings.TrimSpace(pick(matched[0], "rack", "机柜"))
				srcIDC := strings.TrimSpace(rackIDC[srcRack])
				sns := make([]string, 0, len(matched))
				for _, x := range matched {
					sns = append(sns, strings.TrimSpace(pick(x, "sn", "序列号")))
				}
				actions = append(actions, ReconfigActionRow{
					TargetSN:    c.SN,
					GapType:     "硬盘",
					GapQty:      fmt.Sprintf("%d 块", needDiskCount),
					Source:      fmt.Sprintf("库房（机房:%s 机柜:%s）", srcIDC, srcRack),
					PartDetails: strings.Join(sns, ", "),
					CrossIDC:    ternary(srcIDC != "" && c.Datacenter != "" && srcIDC != c.Datacenter, "是", "否"),
					Action:      "改配",
					RuleHit:     "硬盘规格一致（厂商/型号/容量/应用分类）+ 优先库房资源",
				})
			} else {
				donor := s.pickDiskDonor(disks, diskSpec, c.SN, needDiskCount, candidateStatus, serverRack, serverIDC)
				crossIDC := "否"
				if donor.anyIDC != "" && c.Datacenter != "" && donor.anyIDC != c.Datacenter {
					crossIDC = "是"
				}
				partDetails := donor.details
				if donor.missingCount > 0 {
					partDetails = fmt.Sprintf("%s；仍缺 %d 块", donor.details, donor.missingCount)
				}
				actions = append(actions, ReconfigActionRow{
					TargetSN:    c.SN,
					GapType:     "硬盘",
					GapQty:      fmt.Sprintf("%d 块", needDiskCount),
					Source:      donor.source,
					PartDetails: partDetails,
					CrossIDC:    crossIDC,
					Action:      "调拨",
					RuleHit:     "库房不足后按优先级从主机侧调拨",
				})
			}
		}
	}
	return actions
}

type donorPickResult struct {
	source      string
	details     string
	anyIDC      string
	missingCount int
}

func (s *ReconfigService) pickMemoryDonor(memories []map[string]any, memSpec map[string]any, targetSN string, need int, candidateStatus map[string]string, serverRack, serverIDC map[string]string) donorPickResult {
	return pickDonorCommon(memories, targetSN, need, candidateStatus, serverRack, serverIDC,
		func(m map[string]any) bool {
			if strings.TrimSpace(pick(m, "brand", "厂商")) != strings.TrimSpace(pick(memSpec, "brand", "厂商")) {
				return false
			}
			if strings.TrimSpace(pick(m, "model", "型号")) != strings.TrimSpace(pick(memSpec, "model", "型号")) {
				return false
			}
			if pickNum(m, "capacity", "容量") != pickNum(memSpec, "capacity", "容量") {
				return false
			}
			if pickNum(m, "datarate", "数据传输率(TM/s)", "数据传输率(MT/s)") != pickNum(memSpec, "datarate", "数据传输率(TM/s)", "数据传输率(MT/s)") {
				return false
			}
			return true
		},
		func(status string) bool {
			return status != "内存带宽不足"
		},
	)
}

func (s *ReconfigService) pickDiskDonor(disks []map[string]any, diskSpec map[string]any, targetSN string, need int, candidateStatus map[string]string, serverRack, serverIDC map[string]string) donorPickResult {
	return pickDonorCommon(disks, targetSN, need, candidateStatus, serverRack, serverIDC,
		func(d map[string]any) bool {
			if strings.TrimSpace(pick(d, "brand", "厂商")) != strings.TrimSpace(pick(diskSpec, "brand", "厂商")) {
				return false
			}
			if strings.TrimSpace(pick(d, "model", "型号")) != strings.TrimSpace(pick(diskSpec, "model", "型号")) {
				return false
			}
			if pickNum(d, "capacity", "容量") != pickNum(diskSpec, "capacity", "容量") {
				return false
			}
			if strings.TrimSpace(pick(d, "application_category", "应用分类")) != strings.TrimSpace(pick(diskSpec, "application_category", "应用分类")) {
				return false
			}
			return true
		},
		func(status string) bool { return true },
	)
}

func pickDonorCommon(rows []map[string]any, targetSN string, need int, candidateStatus map[string]string, serverRack, serverIDC map[string]string, specMatch func(map[string]any) bool, allowStatus func(string) bool) donorPickResult {
	type part struct {
		partSN   string
		donorSN  string
		donorRack string
		donorIDC string
		priority int
	}
	parts := make([]part, 0)
	for _, r := range rows {
		donorSN := strings.TrimSpace(pick(r, "sn_server", "服务器SN"))
		if donorSN == "" || donorSN == targetSN {
			continue
		}
		if !specMatch(r) {
			continue
		}
		status := candidateStatus[donorSN]
		if !allowStatus(status) {
			continue
		}
		priority := 4
		switch status {
		case "内存带宽不足":
			priority = 1
		case "性能不足":
			priority = 2
		case "候选":
			priority = 3
		}
		parts = append(parts, part{
			partSN:   strings.TrimSpace(pick(r, "sn", "序列号")),
			donorSN:  donorSN,
			donorRack: ternary(serverRack[donorSN] != "", serverRack[donorSN], strings.TrimSpace(pick(r, "rack", "机柜"))),
			donorIDC: serverIDC[donorSN],
			priority: priority,
		})
	}
	sort.Slice(parts, func(i, j int) bool {
		if parts[i].priority != parts[j].priority {
			return parts[i].priority < parts[j].priority
		}
		if parts[i].donorIDC != parts[j].donorIDC {
			return parts[i].donorIDC < parts[j].donorIDC
		}
		if parts[i].donorRack != parts[j].donorRack {
			return parts[i].donorRack < parts[j].donorRack
		}
		return parts[i].donorSN < parts[j].donorSN
	})

	picked := parts
	if len(picked) > need {
		picked = picked[:need]
	}
	if len(picked) == 0 {
		return donorPickResult{
			source:      "无可调拨来源（待人工补货或放宽规则）",
			details:     "规格匹配库存不足，且未找到可调拨主机",
			missingCount: need,
		}
	}
	groups := map[string]struct{}{}
	details := make([]string, 0, len(picked))
	anyIDC := ""
	for _, p := range picked {
		groups[fmt.Sprintf("%s|%s|%s", p.donorSN, p.donorIDC, p.donorRack)] = struct{}{}
		details = append(details, fmt.Sprintf("%s(来源SN:%s 机房:%s 机柜:%s)", p.partSN, p.donorSN, p.donorIDC, p.donorRack))
		if anyIDC == "" {
			anyIDC = p.donorIDC
		}
	}
	sourceList := make([]string, 0, len(groups))
	for g := range groups {
		sourceList = append(sourceList, g)
	}
	sort.Strings(sourceList)
	for i, x := range sourceList {
		arr := strings.Split(x, "|")
		if len(arr) == 3 {
			sourceList[i] = fmt.Sprintf("SN:%s 机房:%s 机柜:%s", arr[0], arr[1], arr[2])
		}
	}
	return donorPickResult{
		source:      "主机调拨（" + strings.Join(sourceList, "；") + "）",
		details:     strings.Join(details, ", "),
		anyIDC:      anyIDC,
		missingCount: maxIntReconfig(0, need-len(picked)),
	}
}

func maxIntReconfig(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func filterBySN(rows []map[string]any, sn string) []map[string]any {
	out := make([]map[string]any, 0)
	for _, r := range rows {
		if strings.TrimSpace(pick(r, "sn_server", "服务器SN")) == sn {
			out = append(out, r)
		}
	}
	return out
}

func sumNum(rows []map[string]any, keys ...string) float64 {
	s := 0.0
	for _, r := range rows {
		s += pickNum(r, keys...)
	}
	return s
}

func pick(row map[string]any, keys ...string) string {
	for _, k := range keys {
		if row == nil {
			continue
		}
		if v, ok := row[k]; ok && v != nil {
			x := strings.TrimSpace(fmt.Sprintf("%v", v))
			if x != "" {
				return x
			}
		}
	}
	return ""
}

func pickNum(row map[string]any, keys ...string) float64 {
	for _, k := range keys {
		if row == nil {
			continue
		}
		if v, ok := row[k]; ok && v != nil {
			s := strings.TrimSpace(fmt.Sprintf("%v", v))
			if s == "" {
				continue
			}
			var n float64
			if _, err := fmt.Sscanf(s, "%f", &n); err == nil {
				return n
			}
		}
	}
	return 0
}

func effectiveMemoryDatarate(hostMems []map[string]any) float64 {
	if len(hostMems) == 0 {
		return 0
	}
	minRate := 0.0
	for i, m := range hostMems {
		rate := pickNum(m, "datarate", "数据传输率(TM/s)", "数据传输率(MT/s)")
		if i == 0 || (rate > 0 && rate < minRate) || minRate <= 0 {
			minRate = rate
		}
	}
	return minRate
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
