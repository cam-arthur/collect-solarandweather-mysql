package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/getinverterdata", getInverterDataHandler).Methods("GET")
	return r
}

func main() {
	r := newRouter()
	err := http.ListenAndServe(":22222", r)
	if err != nil {
		panic(err.Error())
	}
}

func getInverterDataHandler(w http.ResponseWriter, r *http.Request) {
	config := importConfig()
	inverterData := getInverterData(config, runLoginRequest(config))
	weatherData := getWeatherData(config)
	inverter := inverterData.Data.Inverter[0]

	response := ResponseData{
		InverterName:       inverter.Name,
		InverterCapacity:   inverter.Capacity,
		EnergyCurrent:      inverter.D.Pac,
		EnergyDay:          inverter.Eday,
		EnergyMonth:        inverter.Emonth,
		EnergyTotal:        inverter.Etotal,
		LastRead:           inverter.Time,
		OnlineSince:        inverter.TurnonTime,
		CurrentTemperature: weatherData.Main.Temp,
		CloudPercent:       weatherData.Clouds.All,
		WeatherType:        weatherData.Weather[0].Main,
		WeatherDescription: weatherData.Weather[0].Description,
		Sunrise:            weatherData.Sys.Sunrise,
		Sunset:             weatherData.Sys.Sunset,
	}
	responseBytes, err := json.Marshal(response)
	checkErr(err)
	w.Write(responseBytes)
}

func importConfig() Config {
	data, err := os.Open("config.json")
	checkErr(err)
	defer data.Close()
	byteVal, _ := ioutil.ReadAll(data)
	var config Config
	json.Unmarshal(byteVal, &config)
	return config
}

func getWeatherData(config Config) WeatherData {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.WeatherAPI.BaseURL+"zip="+config.WeatherAPI.ZipCode+","+config.WeatherAPI.CountryCode+"&appid="+config.WeatherAPI.AppID+"&units=metric", nil)
	checkErr(err)
	resp, err := client.Do(req)
	checkErr(err)
	defer resp.Body.Close()
	var weatherData WeatherData
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		checkErr(err)
		json.Unmarshal(bodyBytes, &weatherData)
	}
	return weatherData
}

func getInverterData(config Config, LoginResponse LoginResponse) InverterData {
	client := &http.Client{}
	postData := config.ClientConfig.StationInfo
	b, err := json.Marshal(postData)
	checkErr(err)
	req, err := http.NewRequest("POST", config.APIConfig.BaseURL+config.APIConfig.InverterURL, bytes.NewReader(b))
	checkErr(err)
	tokenByte, err := json.Marshal(LoginResponse.Data)
	checkErr(err)
	tokenString := string(tokenByte)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("token", tokenString)
	resp, err := client.Do(req)
	checkErr(err)
	defer resp.Body.Close()
	var inverterData InverterData
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		checkErr(err)
		json.Unmarshal(bodyBytes, &inverterData)
	}
	return inverterData
}

func runLoginRequest(config Config) LoginResponse {
	client := &http.Client{}
	postData := config.ClientConfig.LoginInfo
	b, err := json.Marshal(postData)
	checkErr(err)
	req, err := http.NewRequest("POST", config.APIConfig.BaseURL+config.APIConfig.LoginURL, bytes.NewReader(b))
	checkErr(err)
	tokenByte, err := json.Marshal(config.APIConfig.LoginToken)
	checkErr(err)
	tokenString := string(tokenByte)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", tokenString)
	resp, err := client.Do(req)
	checkErr(err)
	defer resp.Body.Close()
	var loginResponse LoginResponse
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		checkErr(err)
		json.Unmarshal(bodyBytes, &loginResponse)
	}
	return loginResponse
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

type ResponseData struct {
	InverterName       string  `json:"name"`
	InverterCapacity   float64 `json:"capacity"`
	EnergyCurrent      float64 `json:"currentoutput"`
	EnergyDay          float64 `json:"dayoutput"`
	EnergyMonth        float64 `json:"monthOutput"`
	EnergyTotal        float64 `json:"totaloutput"`
	LastRead           string  `json:"readtime"`
	OnlineSince        string  `json:boottime`
	CurrentTemperature float64 `json:"currenttemp"`
	CloudPercent       int     `json:"cloudpercent"`
	WeatherType        string  `json:"weather"`
	WeatherDescription string  `json:"weatherdesc"`
	Sunrise            int     `json:"sunrise"`
	Sunset             int     `json:"sunset"`
}

type ClientConfig struct {
	LoginInfo   LoginInfo   `json:"loginInfo"`
	StationInfo StationInfo `json:"stationInfo"`
}

type StationInfo struct {
	StationID string `json:"powerStationId"`
}

type LoginInfo struct {
	Account  string `json:"account"`
	Password string `json:"pwd"`
}

type APIConfig struct {
	BaseURL     string     `json:"baseURL"`
	LoginURL    string     `json:"loginURL"`
	InverterURL string     `json:"inverterURL"`
	LoginToken  LoginToken `json:"loginToken"`
}

type LoginToken struct {
	Version  string `json:"version"`
	Client   string `json:"client"`
	Language string `json:"language"`
}

type Config struct {
	APIConfig    APIConfig    `json:"apiConfig"`
	ClientConfig ClientConfig `json:"clientConfig"`
	WeatherAPI   WeatherAPI   `json:"weatherAPI"`
}

type WeatherAPI struct {
	BaseURL     string `json:"baseURL"`
	ZipCode     string `json:"zipCode"`
	CountryCode string `json:"countryCode"`
	AppID       string `json:"appid"`
}

type LoginResponse struct {
	HasError bool   `json:"hasError"`
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Data     struct {
		Version   string `json:"version"`
		Client    string `json:"client"`
		Language  string `json:"language"`
		Timestamp int64  `json:"timestamp"`
		UID       string `json:"uid"`
		Token     string `json:"token"`
	} `json:"data"`
	Components struct {
		Para         interface{} `json:"para"`
		LangVer      int         `json:"langVer"`
		TimeSpan     int         `json:"timeSpan"`
		API          string      `json:"api"`
		MsgSocketAdr string      `json:"msgSocketAdr"`
	} `json:"components"`
	API string `json:"api"`
}

type WeatherData struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

type InverterData struct {
	Language string   `json:"language"`
	Function []string `json:"function"`
	HasError bool     `json:"hasError"`
	Msg      string   `json:"msg"`
	Code     string   `json:"code"`
	Data     struct {
		Info struct {
			PowerstationID      string  `json:"powerstation_id"`
			Time                string  `json:"time"`
			DateFormat          string  `json:"date_format"`
			DateFormatYm        string  `json:"date_format_ym"`
			Stationname         string  `json:"stationname"`
			Address             string  `json:"address"`
			OwnerName           string  `json:"owner_name"`
			OwnerPhone          string  `json:"owner_phone"`
			OwnerEmail          string  `json:"owner_email"`
			BatteryCapacity     float64 `json:"battery_capacity"`
			TurnonTime          string  `json:"turnon_time"`
			CreateTime          string  `json:"create_time"`
			Capacity            float64 `json:"capacity"`
			Longitude           float64 `json:"longitude"`
			Latitude            float64 `json:"latitude"`
			PowerstationType    string  `json:"powerstation_type"`
			Status              int     `json:"status"`
			IsStored            bool    `json:"is_stored"`
			IsPowerflow         bool    `json:"is_powerflow"`
			ChartsType          int     `json:"charts_type"`
			HasPv               bool    `json:"has_pv"`
			HasStatisticsCharts bool    `json:"has_statistics_charts"`
			OnlyBps             bool    `json:"only_bps"`
			OnlyBpu             bool    `json:"only_bpu"`
			TimeSpan            float64 `json:"time_span"`
			PrValue             string  `json:"pr_value"`
		} `json:"info"`
		Kpi struct {
			MonthGeneration float64 `json:"month_generation"`
			Pac             float64 `json:"pac"`
			Power           float64 `json:"power"`
			TotalPower      float64 `json:"total_power"`
			DayIncome       float64 `json:"day_income"`
			TotalIncome     float64 `json:"total_income"`
			YieldRate       float64 `json:"yield_rate"`
			Currency        string  `json:"currency"`
		} `json:"kpi"`
		PowercontrolStatus int           `json:"powercontrol_status"`
		Images             []interface{} `json:"images"`
		Weather            struct {
			HeWeather6 []struct {
				DailyForecast []struct {
					CondCodeD string `json:"cond_code_d"`
					CondCodeN string `json:"cond_code_n"`
					CondTxtD  string `json:"cond_txt_d"`
					CondTxtN  string `json:"cond_txt_n"`
					Date      string `json:"date"`
					Time      string `json:"time"`
					Hum       string `json:"hum"`
					Pcpn      string `json:"pcpn"`
					Pop       string `json:"pop"`
					Pres      string `json:"pres"`
					TmpMax    string `json:"tmp_max"`
					TmpMin    string `json:"tmp_min"`
					UvIndex   string `json:"uv_index"`
					Vis       string `json:"vis"`
					WindDeg   string `json:"wind_deg"`
					WindDir   string `json:"wind_dir"`
					WindSc    string `json:"wind_sc"`
					WindSpd   string `json:"wind_spd"`
				} `json:"daily_forecast"`
				Basic struct {
					Cid      string `json:"cid"`
					Location string `json:"location"`
					Cnty     string `json:"cnty"`
					Lat      string `json:"lat"`
					Lon      string `json:"lon"`
					Tz       string `json:"tz"`
				} `json:"basic"`
				Update struct {
					Loc string `json:"loc"`
					Utc string `json:"utc"`
				} `json:"update"`
				Status string `json:"status"`
			} `json:"HeWeather6"`
		} `json:"weather"`
		Inverter []struct {
			Sn   string `json:"sn"`
			Dict struct {
				Left []struct {
					IsHT         bool   `json:"isHT"`
					Key          string `json:"key"`
					Value        string `json:"value"`
					Unit         string `json:"unit"`
					IsFaultMsg   int    `json:"isFaultMsg"`
					FaultMsgCode int    `json:"faultMsgCode"`
				} `json:"left"`
				Right []struct {
					IsHT         bool   `json:"isHT"`
					Key          string `json:"key"`
					Value        string `json:"value"`
					Unit         string `json:"unit"`
					IsFaultMsg   int    `json:"isFaultMsg"`
					FaultMsgCode int    `json:"faultMsgCode"`
				} `json:"right"`
			} `json:"dict"`
			IsStored    bool    `json:"is_stored"`
			Name        string  `json:"name"`
			InPac       float64 `json:"in_pac"`
			OutPac      float64 `json:"out_pac"`
			Eday        float64 `json:"eday"`
			Emonth      float64 `json:"emonth"`
			Etotal      float64 `json:"etotal"`
			Status      int     `json:"status"`
			TurnonTime  string  `json:"turnon_time"`
			ReleationID string  `json:"releation_id"`
			Type        string  `json:"type"`
			Capacity    float64 `json:"capacity"`
			D           struct {
				PwID                  string  `json:"pw_id"`
				Capacity              string  `json:"capacity"`
				Model                 string  `json:"model"`
				OutputPower           string  `json:"output_power"`
				OutputCurrent         string  `json:"output_current"`
				GridVoltage           string  `json:"grid_voltage"`
				BackupOutput          string  `json:"backup_output"`
				Soc                   string  `json:"soc"`
				Soh                   string  `json:"soh"`
				LastRefreshTime       string  `json:"last_refresh_time"`
				WorkMode              string  `json:"work_mode"`
				DcInput1              string  `json:"dc_input1"`
				DcInput2              string  `json:"dc_input2"`
				Battery               string  `json:"battery"`
				BmsStatus             string  `json:"bms_status"`
				Warning               string  `json:"warning"`
				ChargeCurrentLimit    string  `json:"charge_current_limit"`
				DischargeCurrentLimit string  `json:"discharge_current_limit"`
				FirmwareVersion       float64 `json:"firmware_version"`
				CreationDate          string  `json:"creationDate"`
				EDay                  float64 `json:"eDay"`
				ETotal                float64 `json:"eTotal"`
				Pac                   float64 `json:"pac"`
				HTotal                float64 `json:"hTotal"`
				Vpv1                  float64 `json:"vpv1"`
				Vpv2                  float64 `json:"vpv2"`
				Vpv3                  float64 `json:"vpv3"`
				Vpv4                  float64 `json:"vpv4"`
				Ipv1                  float64 `json:"ipv1"`
				Ipv2                  float64 `json:"ipv2"`
				Ipv3                  float64 `json:"ipv3"`
				Ipv4                  float64 `json:"ipv4"`
				Vac1                  float64 `json:"vac1"`
				Vac2                  float64 `json:"vac2"`
				Vac3                  float64 `json:"vac3"`
				Iac1                  float64 `json:"iac1"`
				Iac2                  float64 `json:"iac2"`
				Iac3                  float64 `json:"iac3"`
				Fac1                  float64 `json:"fac1"`
				Fac2                  float64 `json:"fac2"`
				Fac3                  float64 `json:"fac3"`
				Istr1                 float64 `json:"istr1"`
				Istr2                 float64 `json:"istr2"`
				Istr3                 float64 `json:"istr3"`
				Istr4                 float64 `json:"istr4"`
				Istr5                 float64 `json:"istr5"`
				Istr6                 float64 `json:"istr6"`
				Istr7                 float64 `json:"istr7"`
				Istr8                 float64 `json:"istr8"`
				Istr9                 float64 `json:"istr9"`
				Istr10                float64 `json:"istr10"`
				Istr11                float64 `json:"istr11"`
				Istr12                float64 `json:"istr12"`
				Istr13                float64 `json:"istr13"`
				Istr14                float64 `json:"istr14"`
				Istr15                float64 `json:"istr15"`
				Istr16                float64 `json:"istr16"`
			} `json:"d"`
			ItChangeFlag bool        `json:"it_change_flag"`
			Tempperature float64     `json:"tempperature"`
			CheckCode    string      `json:"check_code"`
			Next         interface{} `json:"next"`
			Prev         interface{} `json:"prev"`
			NextDevice   struct {
				Sn       interface{} `json:"sn"`
				IsStored bool        `json:"isStored"`
			} `json:"next_device"`
			PrevDevice struct {
				Sn       interface{} `json:"sn"`
				IsStored bool        `json:"isStored"`
			} `json:"prev_device"`
			InvertFull struct {
				Sn                      string      `json:"sn"`
				PowerstationID          string      `json:"powerstation_id"`
				Name                    string      `json:"name"`
				ModelType               string      `json:"model_type"`
				ChangeType              int         `json:"change_type"`
				ChangeTime              int         `json:"change_time"`
				Capacity                float64     `json:"capacity"`
				Eday                    float64     `json:"eday"`
				Iday                    float64     `json:"iday"`
				Etotal                  float64     `json:"etotal"`
				Itotal                  float64     `json:"itotal"`
				HourTotal               float64     `json:"hour_total"`
				Status                  int         `json:"status"`
				TurnonTime              int64       `json:"turnon_time"`
				Pac                     float64     `json:"pac"`
				Tempperature            float64     `json:"tempperature"`
				Vpv1                    float64     `json:"vpv1"`
				Vpv2                    float64     `json:"vpv2"`
				Vpv3                    float64     `json:"vpv3"`
				Vpv4                    float64     `json:"vpv4"`
				Ipv1                    float64     `json:"ipv1"`
				Ipv2                    float64     `json:"ipv2"`
				Ipv3                    float64     `json:"ipv3"`
				Ipv4                    float64     `json:"ipv4"`
				Vac1                    float64     `json:"vac1"`
				Vac2                    float64     `json:"vac2"`
				Vac3                    float64     `json:"vac3"`
				Iac1                    float64     `json:"iac1"`
				Iac2                    float64     `json:"iac2"`
				Iac3                    float64     `json:"iac3"`
				Fac1                    float64     `json:"fac1"`
				Fac2                    float64     `json:"fac2"`
				Fac3                    float64     `json:"fac3"`
				Istr1                   float64     `json:"istr1"`
				Istr2                   float64     `json:"istr2"`
				Istr3                   float64     `json:"istr3"`
				Istr4                   float64     `json:"istr4"`
				Istr5                   float64     `json:"istr5"`
				Istr6                   float64     `json:"istr6"`
				Istr7                   float64     `json:"istr7"`
				Istr8                   float64     `json:"istr8"`
				Istr9                   float64     `json:"istr9"`
				Istr10                  float64     `json:"istr10"`
				Istr11                  float64     `json:"istr11"`
				Istr12                  float64     `json:"istr12"`
				Istr13                  float64     `json:"istr13"`
				Istr14                  float64     `json:"istr14"`
				Istr15                  float64     `json:"istr15"`
				Istr16                  float64     `json:"istr16"`
				LastTime                int64       `json:"last_time"`
				Vbattery1               float64     `json:"vbattery1"`
				Ibattery1               float64     `json:"ibattery1"`
				Pmeter                  float64     `json:"pmeter"`
				Soc                     float64     `json:"soc"`
				Soh                     float64     `json:"soh"`
				BmsDischargeIMax        interface{} `json:"bms_discharge_i_max"`
				BmsChargeIMax           float64     `json:"bms_charge_i_max"`
				BmsWarning              int         `json:"bms_warning"`
				BmsAlarm                int         `json:"bms_alarm"`
				BattaryWorkMode         int         `json:"battary_work_mode"`
				Workmode                int         `json:"workmode"`
				Vload                   float64     `json:"vload"`
				Iload                   float64     `json:"iload"`
				Firmwareversion         float64     `json:"firmwareversion"`
				Pbackup                 float64     `json:"pbackup"`
				Seller                  float64     `json:"seller"`
				Buy                     float64     `json:"buy"`
				Yesterdaybuytotal       interface{} `json:"yesterdaybuytotal"`
				Yesterdaysellertotal    interface{} `json:"yesterdaysellertotal"`
				Yesterdayct2Sellertotal interface{} `json:"yesterdayct2sellertotal"`
				Yesterdayetotal         interface{} `json:"yesterdayetotal"`
				Yesterdayetotalload     interface{} `json:"yesterdayetotalload"`
				Yesterdaylastime        int         `json:"yesterdaylastime"`
				Thismonthetotle         float64     `json:"thismonthetotle"`
				Lastmonthetotle         float64     `json:"lastmonthetotle"`
				RAM                     float64     `json:"ram"`
				Outputpower             float64     `json:"outputpower"`
				FaultMessge             int         `json:"fault_messge"`
				Isbuettey               bool        `json:"isbuettey"`
				Isbuetteybps            bool        `json:"isbuetteybps"`
				Isbuetteybpu            bool        `json:"isbuetteybpu"`
				IsESUOREMU              bool        `json:"isESUOREMU"`
				BackUpPloadS            float64     `json:"backUpPload_S"`
				BackUpVloadS            float64     `json:"backUpVload_S"`
				BackUpIloadS            float64     `json:"backUpIload_S"`
				BackUpPloadT            float64     `json:"backUpPload_T"`
				BackUpVloadT            float64     `json:"backUpVload_T"`
				BackUpIloadT            float64     `json:"backUpIload_T"`
				ETotalBuy               interface{} `json:"eTotalBuy"`
				EDayBuy                 interface{} `json:"eDayBuy"`
				EBatteryCharge          interface{} `json:"eBatteryCharge"`
				EChargeDay              interface{} `json:"eChargeDay"`
				EBatteryDischarge       interface{} `json:"eBatteryDischarge"`
				EDischargeDay           interface{} `json:"eDischargeDay"`
				BattStrings             float64     `json:"battStrings"`
				MeterConnectStatus      interface{} `json:"meterConnectStatus"`
				MtActivepowerR          float64     `json:"mtActivepowerR"`
				MtActivepowerS          float64     `json:"mtActivepowerS"`
				MtActivepowerT          float64     `json:"mtActivepowerT"`
				EzProConnectStatus      interface{} `json:"ezPro_connect_status"`
				Dataloggersn            string      `json:"dataloggersn"`
				EquipmentName           interface{} `json:"equipment_name"`
				Hasmeter                bool        `json:"hasmeter"`
				MeterType               interface{} `json:"meter_type"`
				PreHourLasttotal        interface{} `json:"pre_hour_lasttotal"`
				PreHourTime             interface{} `json:"pre_hour_time"`
				CurrentHourPv           interface{} `json:"current_hour_pv"`
				ExtendProperties        interface{} `json:"extend_properties"`
				EPConnectStatusHappen   interface{} `json:"eP_connect_status_happen"`
				EPConnectStatusRecover  interface{} `json:"eP_connect_status_recover"`
				TotalSell               float64     `json:"total_sell"`
				TotalBuy                float64     `json:"total_buy"`
				Errors                  interface{} `json:"errors"`
			} `json:"invert_full"`
			Time                     string  `json:"time"`
			Battery                  string  `json:"battery"`
			FirmwareVersion          float64 `json:"firmware_version"`
			WarningBms               string  `json:"warning_bms"`
			Soh                      string  `json:"soh"`
			DischargeCurrentLimitBms string  `json:"discharge_current_limit_bms"`
			ChargeCurrentLimitBms    string  `json:"charge_current_limit_bms"`
			Soc                      string  `json:"soc"`
			PvInput2                 string  `json:"pv_input_2"`
			PvInput1                 string  `json:"pv_input_1"`
			BackUpOutput             string  `json:"back_up_output"`
			OutputVoltage            string  `json:"output_voltage"`
			BackupVoltage            string  `json:"backup_voltage"`
			OutputCurrent            string  `json:"output_current"`
			OutputPower              string  `json:"output_power"`
			TotalGeneration          string  `json:"total_generation"`
			DailyGeneration          string  `json:"daily_generation"`
			BatteryCharging          string  `json:"battery_charging"`
			LastRefreshTime          string  `json:"last_refresh_time"`
			BmsStatus                string  `json:"bms_status"`
			PwID                     string  `json:"pw_id"`
			FaultMessage             string  `json:"fault_message"`
			BatteryPower             float64 `json:"battery_power"`
			PointIndex               string  `json:"point_index"`
			Points                   []struct {
				TargetIndex   int         `json:"target_index"`
				TargetName    string      `json:"target_name"`
				Display       string      `json:"display"`
				Unit          string      `json:"unit"`
				TargetKey     string      `json:"target_key"`
				TextCn        string      `json:"text_cn"`
				TargetSnSix   interface{} `json:"target_sn_six"`
				TargetSnSeven interface{} `json:"target_sn_seven"`
				TargetType    interface{} `json:"target_type"`
				StorageName   interface{} `json:"storage_name"`
			} `json:"points"`
			BackupPloadS       float64     `json:"backup_pload_s"`
			BackupVloadS       float64     `json:"backup_vload_s"`
			BackupIloadS       float64     `json:"backup_iload_s"`
			BackupPloadT       float64     `json:"backup_pload_t"`
			BackupVloadT       float64     `json:"backup_vload_t"`
			BackupIloadT       float64     `json:"backup_iload_t"`
			EtotalBuy          interface{} `json:"etotal_buy"`
			EdayBuy            interface{} `json:"eday_buy"`
			EbatteryCharge     interface{} `json:"ebattery_charge"`
			EchargeDay         interface{} `json:"echarge_day"`
			EbatteryDischarge  interface{} `json:"ebattery_discharge"`
			EdischargeDay      interface{} `json:"edischarge_day"`
			BattStrings        float64     `json:"batt_strings"`
			MeterConnectStatus interface{} `json:"meter_connect_status"`
			MtactivepowerR     float64     `json:"mtactivepower_r"`
			MtactivepowerS     float64     `json:"mtactivepower_s"`
			MtactivepowerT     float64     `json:"mtactivepower_t"`
			HasTigo            bool        `json:"has_tigo"`
			CanStartIV         bool        `json:"canStartIV"`
		} `json:"inverter"`
		Hjgx struct {
			Co2  float64 `json:"co2"`
			Tree float64 `json:"tree"`
			Coal float64 `json:"coal"`
		} `json:"hjgx"`
		PrePowerstationID interface{} `json:"pre_powerstation_id"`
		NexPowerstationID interface{} `json:"nex_powerstation_id"`
		HomKit            struct {
			HomeKitLimit bool        `json:"homeKitLimit"`
			Sn           interface{} `json:"sn"`
		} `json:"homKit"`
		IsTigo                 bool `json:"isTigo"`
		TigoIntervalTimeMinute int  `json:"tigoIntervalTimeMinute"`
		SmuggleInfo            struct {
			IsAllSmuggle    bool        `json:"isAllSmuggle"`
			IsSmuggle       bool        `json:"isSmuggle"`
			DescriptionText interface{} `json:"descriptionText"`
			Sns             interface{} `json:"sns"`
		} `json:"smuggleInfo"`
		HasPowerflow              bool        `json:"hasPowerflow"`
		Powerflow                 interface{} `json:"powerflow"`
		HasEnergeStatisticsCharts bool        `json:"hasEnergeStatisticsCharts"`
		EnergeStatisticsCharts    struct {
			ContributingRate  float64 `json:"contributingRate"`
			SelfUseRate       float64 `json:"selfUseRate"`
			Sum               float64 `json:"sum"`
			Buy               float64 `json:"buy"`
			BuyPercent        float64 `json:"buyPercent"`
			Sell              float64 `json:"sell"`
			SellPercent       float64 `json:"sellPercent"`
			SelfUseOfPv       float64 `json:"selfUseOfPv"`
			ConsumptionOfLoad float64 `json:"consumptionOfLoad"`
			ChartsType        int     `json:"chartsType"`
			HasPv             bool    `json:"hasPv"`
			HasCharge         bool    `json:"hasCharge"`
			Charge            float64 `json:"charge"`
			DisCharge         float64 `json:"disCharge"`
		} `json:"energeStatisticsCharts"`
		EnergeStatisticsTotals struct {
			ContributingRate  float64 `json:"contributingRate"`
			SelfUseRate       float64 `json:"selfUseRate"`
			Sum               float64 `json:"sum"`
			Buy               float64 `json:"buy"`
			BuyPercent        float64 `json:"buyPercent"`
			Sell              float64 `json:"sell"`
			SellPercent       float64 `json:"sellPercent"`
			SelfUseOfPv       float64 `json:"selfUseOfPv"`
			ConsumptionOfLoad float64 `json:"consumptionOfLoad"`
			ChartsType        int     `json:"chartsType"`
			HasPv             bool    `json:"hasPv"`
			HasCharge         bool    `json:"hasCharge"`
			Charge            float64 `json:"charge"`
			DisCharge         float64 `json:"disCharge"`
		} `json:"energeStatisticsTotals"`
		Soc struct {
			Power  int `json:"power"`
			Status int `json:"status"`
		} `json:"soc"`
		Environmental []interface{} `json:"environmental"`
		Equipment     []struct {
			Type                 string      `json:"type"`
			Title                string      `json:"title"`
			Status               int         `json:"status"`
			StatusText           interface{} `json:"statusText"`
			Capacity             interface{} `json:"capacity"`
			ActionThreshold      interface{} `json:"actionThreshold"`
			SubordinateEquipment string      `json:"subordinateEquipment"`
			PowerGeneration      string      `json:"powerGeneration"`
			Eday                 string      `json:"eday"`
			Brand                string      `json:"brand"`
			IsStored             bool        `json:"isStored"`
			Soc                  string      `json:"soc"`
			IsChange             bool        `json:"isChange"`
			RelationID           string      `json:"relationId"`
			Sn                   string      `json:"sn"`
			HasTigo              bool        `json:"has_tigo"`
			IsSec                bool        `json:"is_sec"`
			IsSecs               bool        `json:"is_secs"`
			TargetPF             interface{} `json:"targetPF"`
			ExportPowerlimit     interface{} `json:"exportPowerlimit"`
		} `json:"equipment"`
	} `json:"data"`
	Components struct {
		Para         string      `json:"para"`
		LangVer      int         `json:"langVer"`
		TimeSpan     int         `json:"timeSpan"`
		API          string      `json:"api"`
		MsgSocketAdr interface{} `json:"msgSocketAdr"`
	} `json:"components"`
}
