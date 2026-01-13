import React from 'react';
import ReactDOM from 'react-dom/client';
import { ConfigProvider, theme as antdTheme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import App from './App';
import { useThemeStore } from './store/themeStore';
import './index.css';

function Root() {
  const { theme: themeMode } = useThemeStore();
  
  // 设置根元素的 data-theme 属性，用于 CSS 主题检测
  React.useEffect(() => {
    document.documentElement.setAttribute('data-theme', themeMode);
  }, [themeMode]);
  
  // 自定义暗色主题配置
  const darkThemeConfig = {
    token: {
      // 主色调 - 使用更柔和的蓝色
      colorPrimary: '#4A9EFF',
      colorSuccess: '#52C41A',
      colorWarning: '#FAAD14',
      colorError: '#FF4D4F',
      colorInfo: '#4A9EFF',
      
      // 背景色 - 使用深灰而非纯黑，更有层次感
      colorBgBase: '#0F1419',
      colorBgContainer: '#1A1F2E',
      colorBgElevated: '#252B3A',
      colorBgLayout: '#0F1419',
      
      // 文字颜色 - 提高对比度
      colorText: 'rgba(255, 255, 255, 0.95)',
      colorTextSecondary: 'rgba(255, 255, 255, 0.75)',
      colorTextTertiary: 'rgba(255, 255, 255, 0.55)',
      colorTextQuaternary: 'rgba(255, 255, 255, 0.35)',
      
      // 边框颜色 - 更柔和的边框
      colorBorder: 'rgba(255, 255, 255, 0.12)',
      colorBorderSecondary: 'rgba(255, 255, 255, 0.08)',
      
      // 阴影 - 更柔和的阴影效果
      boxShadow: '0 2px 8px rgba(0, 0, 0, 0.3)',
      boxShadowSecondary: '0 1px 4px rgba(0, 0, 0, 0.2)',
      
      // 圆角
      borderRadius: 8,
      borderRadiusLG: 12,
      borderRadiusSM: 6,
      
      // 字体
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
    },
    algorithm: antdTheme.darkAlgorithm,
  };
  
  return (
    <ConfigProvider
      locale={zhCN}
      theme={themeMode === 'dark' ? darkThemeConfig : {
        token: {
          borderRadius: 8,
          borderRadiusLG: 12,
          borderRadiusSM: 6,
        },
      }}
    >
      <App />
    </ConfigProvider>
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <Root />
  </React.StrictMode>
);
