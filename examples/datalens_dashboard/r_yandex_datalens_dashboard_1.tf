//
// Create a DataLens dashboard. The payload mirrors the DataLens API:
// `entry { annotation, meta, data }` carries the dashboard content.
//
resource "yandex_datalens_dashboard" "my-dashboard" {
  organization_id = "example-organization-id"

  entry = {
    name        = "example-dashboard"
    workbook_id = "example-workbook-id"
    annotation = {
      description = "Sales overview"
    }
    meta = {
      title  = "Sales overview"
      locale = "en"
    }
    data = {
      counter = 1
      salt    = "abcde"

      settings = {
        autoupdate_interval = 0
        hide_tabs           = false
      }

      tabs = [
        {
          id    = "tab-1"
          title = "Overview"
          items = [
            {
              id = "widget-1"
              widget = {
                tabs = [{
                  id       = "wt-1"
                  title    = "Chart"
                  chart_id = yandex_datalens_chart.my-chart.id
                }]
              }
            },
          ]
          layout = [
            { i = "widget-1", h = 12, w = 12, x = 0, y = 0 },
          ]
          connections = []
        },
      ]
    }
  }
}
