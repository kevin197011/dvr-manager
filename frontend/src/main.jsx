import React from 'react';
import ReactDOM from 'react-dom/client';
import { ConfigProvider, theme as antdTheme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import App from './App';
import { useThemeStore } from './store/themeStore';
import { getAntdTheme, getCSSVariables } from './config/theme';
import './index.css';

function Root() {
  const { theme: themeMode } = useThemeStore();
  
  // 设置根元素的 data-theme 属性，用于 CSS 主题检测
  React.useEffect(() => {
    document.documentElement.setAttribute('data-theme', themeMode);
    
    // 应用 CSS 变量
    const cssVars = getCSSVariables(themeMode);
    const root = document.documentElement;
    Object.entries(cssVars).forEach(([key, value]) => {
      root.style.setProperty(key, value);
    });
  }, [themeMode]);
  
  // 获取 Ant Design 主题配置
  const antdThemeConfig = getAntdTheme(themeMode);
  
  // 如果是暗色模式，应用 darkAlgorithm
  if (themeMode === 'dark') {
    antdThemeConfig.algorithm = antdTheme.darkAlgorithm;
  }
  
  return (
    <ConfigProvider
      locale={zhCN}
      theme={antdThemeConfig}
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
