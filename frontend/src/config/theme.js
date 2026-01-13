/**
 * 统一设计系统配置
 * 基于 Minimalism + Dark Mode (OLED) 风格
 * 参考：Analytics Dashboard 配色方案
 */

export const designSystem = {
  // 颜色系统
  colors: {
    // 主色调 - Analytics Dashboard 配色
    primary: {
      main: '#3B82F6',      // 主色
      light: '#60A5FA',      // 浅色
      dark: '#2563EB',       // 深色
      hover: '#2563EB',      // 悬停
      active: '#1D4ED8',     // 激活
    },
    
    // 功能色
    success: '#22C55E',
    warning: '#F97316',
    error: '#EF4444',
    info: '#3B82F6',
    
    // 浅色模式
    light: {
      background: {
        base: '#FFFFFF',
        container: '#F8FAFC',
        elevated: '#FFFFFF',
        layout: '#F8FAFC',
      },
      text: {
        primary: '#1E293B',
        secondary: '#64748B',
        tertiary: '#94A3B8',
        disabled: '#CBD5E1',
      },
      border: {
        primary: '#E2E8F0',
        secondary: '#F1F5F9',
        tertiary: '#F8FAFC',
      },
      shadow: {
        sm: '0 1px 2px rgba(0, 0, 0, 0.05)',
        md: '0 4px 12px rgba(0, 0, 0, 0.08)',
        lg: '0 8px 24px rgba(0, 0, 0, 0.12)',
      },
    },
    
    // 暗色模式 (OLED)
    dark: {
      background: {
        base: '#0F172A',      // Deep Black
        container: '#1E293B',  // Dark Grey
        elevated: '#334155',   // Elevated
        layout: '#0F172A',     // Layout
      },
      text: {
        primary: '#F8FAFC',    // 高对比度白色
        secondary: '#CBD5E1',  // 次要文本
        tertiary: '#94A3B8',   // 三级文本
        disabled: '#64748B',   // 禁用文本
      },
      border: {
        primary: 'rgba(255, 255, 255, 0.12)',
        secondary: 'rgba(255, 255, 255, 0.08)',
        tertiary: 'rgba(255, 255, 255, 0.05)',
      },
      shadow: {
        sm: '0 1px 3px rgba(0, 0, 0, 0.3)',
        md: '0 4px 16px rgba(0, 0, 0, 0.4)',
        lg: '0 8px 32px rgba(0, 0, 0, 0.5)',
      },
    },
  },
  
  // 间距系统（8px 基础单位）
  spacing: {
    xs: '4px',
    sm: '8px',
    md: '16px',
    lg: '24px',
    xl: '32px',
    xxl: '48px',
  },
  
  // 圆角系统
  borderRadius: {
    sm: '6px',
    md: '8px',
    lg: '12px',
    xl: '16px',
    full: '9999px',
  },
  
  // 字体系统
  typography: {
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
    fontSize: {
      xs: '12px',
      sm: '14px',
      md: '16px',
      lg: '18px',
      xl: '20px',
      '2xl': '24px',
      '3xl': '32px',
    },
    fontWeight: {
      normal: 400,
      medium: 500,
      semibold: 600,
      bold: 700,
    },
    lineHeight: {
      tight: 1.2,
      normal: 1.5,
      relaxed: 1.75,
    },
  },
  
  // 过渡动画
  transition: {
    fast: '150ms',
    normal: '200ms',
    slow: '300ms',
    easing: 'cubic-bezier(0.4, 0, 0.2, 1)',
  },
  
  // Z-index 层级
  zIndex: {
    dropdown: 1000,
    sticky: 1020,
    fixed: 1030,
    modalBackdrop: 1040,
    modal: 1050,
    popover: 1060,
    tooltip: 1070,
  },
};

/**
 * 获取 Ant Design 主题配置
 */
export const getAntdTheme = (themeMode = 'light') => {
  const isDark = themeMode === 'dark';
  const colors = isDark ? designSystem.colors.dark : designSystem.colors.light;
  
  // 动态导入 antd theme
  let darkAlgorithm = undefined;
  if (isDark) {
    // 在运行时动态导入
    import('antd').then((antd) => {
      darkAlgorithm = antd.theme.darkAlgorithm;
    });
  }
  
  return {
    token: {
      // 主色调
      colorPrimary: designSystem.colors.primary.main,
      colorSuccess: designSystem.colors.success,
      colorWarning: designSystem.colors.warning,
      colorError: designSystem.colors.error,
      colorInfo: designSystem.colors.info,
      
      // 背景色
      colorBgBase: colors.background.base,
      colorBgContainer: colors.background.container,
      colorBgElevated: colors.background.elevated,
      colorBgLayout: colors.background.layout,
      
      // 文字颜色
      colorText: colors.text.primary,
      colorTextSecondary: colors.text.secondary,
      colorTextTertiary: colors.text.tertiary,
      colorTextQuaternary: colors.text.disabled,
      
      // 边框颜色
      colorBorder: colors.border.primary,
      colorBorderSecondary: colors.border.secondary,
      
      // 圆角
      borderRadius: parseInt(designSystem.borderRadius.md),
      borderRadiusLG: parseInt(designSystem.borderRadius.lg),
      borderRadiusSM: parseInt(designSystem.borderRadius.sm),
      
      // 字体
      fontFamily: designSystem.typography.fontFamily,
      
      // 阴影
      boxShadow: colors.shadow.md,
      boxShadowSecondary: colors.shadow.sm,
    },
    // algorithm 将在 main.jsx 中根据 themeMode 设置
    algorithm: undefined,
  };
};

/**
 * 获取 CSS 变量（用于纯 CSS 样式）
 */
export const getCSSVariables = (themeMode = 'light') => {
  const isDark = themeMode === 'dark';
  const colors = isDark ? designSystem.colors.dark : designSystem.colors.light;
  
  return {
    '--color-primary': designSystem.colors.primary.main,
    '--color-primary-hover': designSystem.colors.primary.hover,
    '--color-primary-active': designSystem.colors.primary.active,
    
    '--color-success': designSystem.colors.success,
    '--color-warning': designSystem.colors.warning,
    '--color-error': designSystem.colors.error,
    '--color-info': designSystem.colors.info,
    
    '--bg-base': colors.background.base,
    '--bg-container': colors.background.container,
    '--bg-elevated': colors.background.elevated,
    '--bg-layout': colors.background.layout,
    
    '--text-primary': colors.text.primary,
    '--text-secondary': colors.text.secondary,
    '--text-tertiary': colors.text.tertiary,
    '--text-disabled': colors.text.disabled,
    
    '--border-primary': colors.border.primary,
    '--border-secondary': colors.border.secondary,
    '--border-tertiary': colors.border.tertiary,
    
    '--shadow-sm': colors.shadow.sm,
    '--shadow-md': colors.shadow.md,
    '--shadow-lg': colors.shadow.lg,
    
    '--spacing-xs': designSystem.spacing.xs,
    '--spacing-sm': designSystem.spacing.sm,
    '--spacing-md': designSystem.spacing.md,
    '--spacing-lg': designSystem.spacing.lg,
    '--spacing-xl': designSystem.spacing.xl,
    '--spacing-xxl': designSystem.spacing.xxl,
    
    '--radius-sm': designSystem.borderRadius.sm,
    '--radius-md': designSystem.borderRadius.md,
    '--radius-lg': designSystem.borderRadius.lg,
    '--radius-xl': designSystem.borderRadius.xl,
    
    '--transition-fast': designSystem.transition.fast,
    '--transition-normal': designSystem.transition.normal,
    '--transition-slow': designSystem.transition.slow,
    '--transition-easing': designSystem.transition.easing,
  };
};

export default designSystem;
