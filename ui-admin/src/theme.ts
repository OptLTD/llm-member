import type { GlobalThemeOverrides } from 'naive-ui'

export const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#6366f1',
    primaryColorHover: '#4f46e5',
    primaryColorPressed: '#4338ca',
    borderRadius: '4px',
    fontFamily: "system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif",
  },
  Card: {
    borderRadius: '8px',
    paddingSmall: '8px 10px',
    paddingMedium: '12px 14px',
    paddingLarge: '14px 16px',
  },
  DataTable: {
    borderRadius: '8px',
  },
}
