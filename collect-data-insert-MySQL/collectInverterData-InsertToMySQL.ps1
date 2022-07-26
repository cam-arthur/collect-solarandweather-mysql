function ConvertToMySQLDatetime($datetime){
    try {
        return [datetime]::parseexact($datetime, 'MM/dd/yyyy HH:mm:ss', $null).ToString('yyyy-MM-dd HH:mm:ss')
    }catch{
        return ([datetime]'1970-01-01 00:00:00').AddSeconds($datetime).ToLocalTime().ToString('yyyy-MM-dd HH:mm:ss')
    }
}

Add-Type -Path 'C:\Program Files (x86)\MySQL\Connector NET 8.0\Assemblies\v4.5.2\MySql.Data.dll'


while ($true){
    $response = Invoke-RestMethod -Method "GET" -Uri "http://API/getinverterdata" 

    $response.readtime = ConvertToMySQLDatetime $response.readtime
    $response.boottime = ConvertToMySQLDatetime $response.boottime
    $response.sunrise = ConvertToMySQLDatetime $response.sunrise
    $response.sunset = ConvertToMySQLDatetime $response.sunset

    $sqlQuery = "insert into inverter_data (inverter_name, inverter_capacity, inverter_current, inverter_day_total, inverter_month_total, inverter_total, read_time, boot_time, current_temp, cloud_percent, weather, weather_description, sunrise, sunset) values ("
    $sqlQuery += "`"$($response.name)`", $($response.capacity), $($response.currentoutput), $($response.dayoutput), $($response.monthOutput), $($response.totaloutput), `"$($response.readtime)`", `"$($response.boottime)`", $($response.currenttemp), $($response.cloudpercent), `"$($response.weather)`", `"$($response.weatherdesc)`", `"$($response.sunrise)`", `"$($response.sunset)`")"

    $Connection = [MySql.Data.MySqlClient.MySqlConnection]@{ConnectionString='server=localhost;uid=;pwd=;database='}
    $Connection.Open()
 
    # Define a MySQL Command Object for a non-query.
    $sql = New-Object MySql.Data.MySqlClient.MySqlCommand
    $sql.Connection = $Connection
    $sql.CommandText = $sqlQuery
    $sql.ExecuteNonQuery()
 
    # Close the MySQL connection.
    $Connection.Close()

    sleep 60
}