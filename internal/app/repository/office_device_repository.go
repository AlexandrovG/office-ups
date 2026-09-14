package repository

import "errors"

type OfficeDevice struct {
	ID                int
	DeviceName        string
	DeviceDescription string
	ActivePowerWatts  int
	PowerFactor       float64
	DeviceStatus      string
	LikedUserIDs      []int
	ImageKey          string
	VideoKey          string
}

type OfficeDeviceRepository struct {
	devices []OfficeDevice
}

func NewOfficeDeviceRepository() *OfficeDeviceRepository {
	return &OfficeDeviceRepository{
		devices: []OfficeDevice{
			{
				ID:                1,
				DeviceName:        "Рабочая станция инженера",
				DeviceDescription: "Системный блок Core i7, 32GB RAM, БП 750W с APFC. Высокие пусковые токи.",
				ActivePowerWatts:  450,
				PowerFactor:       0.85,
				DeviceStatus:      "PUBLISHED",
				LikedUserIDs:      []int{101, 102, 105, 107},
				ImageKey:          "workstation.jpg",
				VideoKey:          "workstation_preview.mp4",
			},
			{
				ID:                2,
				DeviceName:        "Монитор оператора 27 IPS",
				DeviceDescription: "Дисплей графического поста системного энергомониторинга.",
				ActivePowerWatts:  45,
				PowerFactor:       0.95,
				DeviceStatus:      "PUBLISHED",
				LikedUserIDs:      []int{101, 108},
				ImageKey:          "monitor27.jpg",
				VideoKey:          "workstation_preview2.mp4",
			},
			{
				ID:                3,
				DeviceName:        "Маршрутизатор ядра MikroTik",
				DeviceDescription: "Сетевой шлюз доступа офиса. Критичен к непрерывности электропитания.",
				ActivePowerWatts:  60,
				PowerFactor:       0.70,
				DeviceStatus:      "PUBLISHED",
				LikedUserIDs:      []int{103, 104, 109, 110, 111},
				ImageKey:          "router.jpg",
				VideoKey:          "workstation_preview.mp4",
			},
			{
				ID:                4,
				DeviceName:        "Сетевое хранилище NAS 4-Bay",
				DeviceDescription: "Резервный накопитель БД. Требует штатного отключения по USB-сигналу ИБП.",
				ActivePowerWatts:  120,
				PowerFactor:       0.80,
				DeviceStatus:      "PUBLISHED",
				LikedUserIDs:      []int{102},
				ImageKey:          "nas.jpg",
				VideoKey:          "workstation_preview.mp4",
			},
			{
				ID:                5,
				DeviceName:        "Лазерное МФУ рабочей группы",
				DeviceDescription: "Импульсная пиковая нагрузка при периодическом включении нагревателя фьюзера.",
				ActivePowerWatts:  850,
				PowerFactor:       0.65,
				DeviceStatus:      "DRAFT",
				LikedUserIDs:      []int{},
				ImageKey:          "printer_draft.jpg",
				VideoKey:          "workstation_preview.mp4",
			},
			{
				ID:                6,
				DeviceName:        "Списанный сетевой хаб 100M",
				DeviceDescription: "Устаревшее оборудование",
				ActivePowerWatts:  15,
				PowerFactor:       0.60,
				DeviceStatus:      "DELETED",
				LikedUserIDs:      []int{},
				ImageKey:          "old_hub.jpg",
				VideoKey:          "workstation_preview.mp4",
			},
		},
	}
}

func (r *OfficeDeviceRepository) GetPublishedDevices(minWatts int) []OfficeDevice {
	var result []OfficeDevice
	for _, d := range r.devices {
		if d.DeviceStatus == "PUBLISHED" && d.ActivePowerWatts >= minWatts {
			result = append(result, d)
		}
	}
	return result
}

func (r *OfficeDeviceRepository) GetDeviceByID(id int) (OfficeDevice, error) {
	for _, d := range r.devices {
		if d.ID == id && d.DeviceStatus != "DELETED" {
			return d, nil
		}
	}
	return OfficeDevice{}, errors.New("device not found")
}

func (r *OfficeDeviceRepository) GetNextPublishedDevice(currentID int) OfficeDevice {
	var published []OfficeDevice
	for _, d := range r.devices {
		if d.DeviceStatus == "PUBLISHED" {
			published = append(published, d)
		}
	}
	if len(published) == 0 {
		return OfficeDevice{}
	}
	for i, d := range published {
		if d.ID == currentID {
			if i+1 < len(published) {
				return published[i+1]
			}
			return published[0]
		}
	}
	return published[0]
}

func (r *OfficeDeviceRepository) GetDraftDevice() (OfficeDevice, error) {
	for _, d := range r.devices {
		if d.DeviceStatus == "DRAFT" {
			return d, nil
		}
	}
	return OfficeDevice{}, errors.New("draft not found")
}