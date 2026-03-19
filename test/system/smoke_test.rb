require "application_system_test_case"

class SmokeTest < ApplicationSystemTestCase
  test "visiting dashboard works" do
    visit root_url

    assert_selector ".dashboard-container"
  end
end
