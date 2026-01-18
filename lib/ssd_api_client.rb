#!/usr/bin/env ruby
# frozen_string_literal: true

require "net/http"
require "uri"
require "json"
require "date"

class SsdApiClient
  BASE_URL = "https://ims.ssd.sk"

  def initialize
    @username = ENV.fetch("IMS_SSD_USERNAME") { raise "IMS_SSD_USERNAME environment variable is required" }
    @password = ENV.fetch("IMS_SSD_PASSWORD") { raise "IMS_SSD_PASSWORD environment variable is required" }
    @pointOfDeliveryId = ENV.fetch("IMS_SSD_POINT_OF_DELIVERY_ID") { raise "IMS_SSD_POINT_OF_DELIVERY_ID environment variable is required" }
    @cookies = {}
  end

  def fetch_profile_data_for_date(date = Date.today - 1)
    @target_date = date.is_a?(Date) ? date : Date.parse(date.to_s)
    login
    fetch_profile_data
  end

  private

  def login
    uri = URI("#{BASE_URL}/api/account/login")

    request = Net::HTTP::Post.new(uri)
    request["Accept"] = "application/json"
    request["Content-Type"] = "application/json"
    request["Origin"] = BASE_URL
    request["Referer"] = "#{BASE_URL}/login"
    request["X-Requested-With"] = "XMLHttpRequest"

    request.body = {
      username: @username,
      password: @password
    }.to_json

    response = make_request(uri, request)

    # Extract cookies from response
    extract_cookies(response)

    puts "Login successful!"
    response
  end

  def fetch_profile_data
    uri = URI("#{BASE_URL}/api/consumption-production/profile-data")

    request = Net::HTTP::Post.new(uri)
    request["Accept"] = "application/json"
    request["Content-Type"] = "application/json"
    request["Origin"] = BASE_URL
    request["Referer"] = "#{BASE_URL}/consumption-production/profile-data"
    request["X-Requested-With"] = "XMLHttpRequest"
    request["X-Client-Screen-Number"] = "CP.015"
    request["X-Client-Url"] = "/consumption-production/profile-data"
    request["Cookie"] = cookie_header

    request.body = build_profile_data_request_body.to_json

    response = make_request(uri, request)

    json_response = JSON.parse(response.body)
    json_response["rows"].map { |row| { incoming: row["values"][2], outgoing: row["values"][4] } }
  end

  def build_profile_data_request_body
    # Calculate date range (UTC) for the target date
    day_start = @target_date.to_time.utc
    day_end = (@target_date + 1).to_time.utc

    {
      page: {
        totalRows: 0,
        currentPage: 1,
        pageSize: 100
      },
      columns: [
        { member: "meteringDatetime", title: "Dátum a čas merania", type: "DateTime", isVisible: true, isSortable: true, isFilterable: false, isNullable: false, index: 0, allowVisible: true },
        { member: "period", title: "Perióda", type: "Int", isVisible: true, isSortable: true, isFilterable: true, isNullable: false, index: 1, allowVisible: true },
        { member: "actualConsumption", title: "1.5.0 - Činný odber (kW)", type: "Float", isVisible: true, isSortable: true, isFilterable: true, isNullable: true, index: 2, allowVisible: true },
        { member: "actualConsumptionQualityType", title: "1.5.0 - Činný odber (kvalita)", type: "Enumeration", isVisible: true, isSortable: true, isFilterable: true, isNullable: true, index: 3, allowVisible: true },
        { member: "actualSupply", title: "2.5.0 - Činná dodávka (kW)", type: "Float", isVisible: true, isSortable: true, isFilterable: true, isNullable: true, index: 4, allowVisible: true },
        { member: "actualSupplyQualityType", title: "2.5.0 - Činná dodávka (kvalita)", type: "Enumeration", isVisible: true, isSortable: true, isFilterable: true, isNullable: true, index: 5, allowVisible: true },
        { member: "idleConsumption", title: "3.5.0 - Jalový odber (kVAR)", type: "Float", isVisible: true, isSortable: true, isFilterable: true, isNullable: true, index: 6, allowVisible: true },
        { member: "idleConsumptionQualityType", title: "3.5.0 - Jalový odber (kvalita)", type: "Enumeration", isVisible: true, isSortable: true, isFilterable: true, isNullable: true, index: 7, allowVisible: true },
        { member: "idleSupply", title: "4.5.0 - Jalová dodávka (kVAR)", type: "Float", isVisible: true, isSortable: true, isFilterable: true, isNullable: true, index: 8, allowVisible: true },
        { member: "idleSupplyQualityType", title: "4.5.0 - Jalová dodávka (kvalita)", type: "Enumeration", isVisible: true, isSortable: true, isFilterable: true, isNullable: true, index: 9, allowVisible: true },
        { member: "profileDataId", type: "UNDEFINED", isVisible: false, isSortable: false, isFilterable: false, isNullable: false, index: 10, allowVisible: false },
        { member: "pointOfDeliveryId", type: "UNDEFINED", isVisible: false, isSortable: false, isFilterable: false, isNullable: false, index: 11, allowVisible: false },
        { member: "startPeriodDatetime", type: "UNDEFINED", isVisible: false, isSortable: false, isFilterable: false, isNullable: false, index: 12, allowVisible: false },
        { member: "userId", type: "UNDEFINED", isVisible: false, isSortable: false, isFilterable: false, isNullable: false, index: 13, allowVisible: false },
        { member: "changedOn", type: "UNDEFINED", isVisible: false, isSortable: false, isFilterable: false, isNullable: false, index: 14, allowVisible: false },
        { member: "deletedOn", type: "UNDEFINED", isVisible: false, isSortable: false, isFilterable: false, isNullable: false, index: 15, allowVisible: false }
      ],
      filters: [
        {
          member: "pointOfDeliveryId",
          operator: "Equals",
          type: "Int",
          value: @pointOfDeliveryId
        },
        {
          member: "meteringDatetime",
          operator: "Greater",
          type: "DateTimeMilliseconds",
          value: day_start.strftime("%Y-%m-%dT%H:%M:%S.000Z"),
          rangeOperator: "LowerOrEquals",
          rangeValue: day_end.strftime("%Y-%m-%dT%H:%M:%S.000Z")
        },
        {
          member: "meteringDatetime",
          operator: "LowerOrEquals",
          type: "DateTimeMilliseconds",
          value: day_end.strftime("%Y-%m-%dT%H:%M:%S.000Z"),
          rangeOperator: nil,
          rangeValue: nil
        }
      ],
      sort: [
        { member: "meteringDatetime", sortOrder: "asc" }
      ],
      isExport: false
    }
  end

  def make_request(uri, request)
    http = Net::HTTP.new(uri.host, uri.port)
    http.use_ssl = true
    http.verify_mode = OpenSSL::SSL::VERIFY_PEER

    response = http.request(request)

    unless response.is_a?(Net::HTTPSuccess)
      raise "HTTP Error: #{response.code} - #{response.message}\nBody: #{response.body}"
    end

    response
  end

  def extract_cookies(response)
    response.get_fields("Set-Cookie")&.each do |cookie_str|
      cookie_parts = cookie_str.split(";").first
      name, value = cookie_parts.split("=", 2)
      @cookies[name] = value if name && value
    end
  end

  def cookie_header
    @cookies.map { |name, value| "#{name}=#{value}" }.join("; ")
  end
end

# Run the script when executed directly
if __FILE__ == $PROGRAM_NAME
  begin
    client = SsdApiClient.new
    response = client.fetch_yesterday_profile_data
    puts "\nProfile data fetched successfully!"
  rescue StandardError => e
    puts "Error: #{e.message}"
    exit 1
  end
end
