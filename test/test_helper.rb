ENV["RAILS_ENV"] ||= "test"
require_relative "../config/environment"
require "rails/test_help"

module ActiveSupport
  class TestCase
    # Run tests in parallel with specified workers
    parallelize(workers: :number_of_processors)

    # Setup all fixtures in test/fixtures/*.yml for all tests in alphabetical order.
    fixtures :all

    # Add more helper methods to be used by all tests here...
  end
end

class FakeSsdApiClient
  attr_reader :requested_dates

  def initialize(responder)
    @responder = responder
    @requested_dates = []
  end

  def fetch_profile_data_for_date(date)
    @requested_dates << date
    @responder.call(date)
  end
end

module SsdApiClientStubbing
  # Replaces SsdApiClient.new with a fake; the block receives a date and returns profile rows.
  def stub_ssd_api_client(&responder)
    client = FakeSsdApiClient.new(responder || ->(_date) { [] })
    unstub_ssd_api_client
    SsdApiClient.define_singleton_method(:new) { |*| client }
    client
  end

  def unstub_ssd_api_client
    singleton = SsdApiClient.singleton_class
    singleton.remove_method(:new) if singleton.instance_methods(false).include?(:new)
  end
end
