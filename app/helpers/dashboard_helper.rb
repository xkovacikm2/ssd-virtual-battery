module DashboardHelper
  SLOVAK_DAY_NAMES = %w[Nedeľa Pondelok Utorok Streda Štvrtok Piatok Sobota].freeze
  SLOVAK_MONTH_NAMES = [ nil, "Január", "Február", "Marec", "Apríl", "Máj", "Jún",
                         "Júl", "August", "September", "Október", "November", "December" ].freeze

  def slovak_day_name(date)
    SLOVAK_DAY_NAMES[date.wday]
  end

  def slovak_month_name(month)
    SLOVAK_MONTH_NAMES[month]
  end
end
