module MetricsFormatter
  def self.call(data)
    data.to_json
  end
end
