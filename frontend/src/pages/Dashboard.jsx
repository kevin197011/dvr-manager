import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  DatePicker,
  Button,
  Space,
  message,
  Typography,
} from 'antd';
import { Line } from '@ant-design/plots';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { adminService } from '../services/authService';
import { useThemeStore } from '../store/themeStore';

const { RangePicker } = DatePicker;
const { Text } = Typography;

function pct(rate) {
  if (rate == null || Number.isNaN(rate)) return '—';
  return `${Math.round(rate * 100)}%`;
}

function defaultRange() {
  return [dayjs().subtract(29, 'day'), dayjs()];
}

function Dashboard() {
  const theme = useThemeStore((s) => s.theme);
  const [range, setRange] = useState(defaultRange);
  const [loading, setLoading] = useState(false);
  const [summary, setSummary] = useState(null);
  const [series, setSeries] = useState([]);
  const [truncated, setTruncated] = useState(false);

  const fetchStats = useCallback(async () => {
    if (!range?.[0] || !range?.[1]) return;
    setLoading(true);
    try {
      const res = await adminService.getDashboardStats({
        from: range[0].format('YYYY-MM-DD'),
        to: range[1].format('YYYY-MM-DD'),
      });
      if (res?.success) {
        setSummary(res.summary || {});
        setSeries(res.series || []);
        setTruncated(!!res.truncated);
      } else {
        message.error(res?.message || '获取统计数据失败');
      }
    } catch (err) {
      message.error(err?.response?.data?.message || '获取统计数据失败');
    } finally {
      setLoading(false);
    }
  }, [range]);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  const chartData = useMemo(() => {
    const types = [
      { key: 'query_single', label: '单次查询' },
      { key: 'query_batch', label: '批量查询' },
      { key: 'stream', label: '流访问' },
    ];
    return series.flatMap((row) =>
      types.map(({ key, label }) => ({
        date: row.date,
        value: row[key] ?? 0,
        type: label,
      }))
    );
  }, [series]);

  const lineConfig = useMemo(
    () => ({
      data: chartData,
      xField: 'date',
      yField: 'value',
      seriesField: 'type',
      smooth: true,
      height: 320,
      theme: theme === 'dark' ? 'dark' : 'light',
      legend: { position: 'top' },
      xAxis: { tickCount: Math.min(series.length, 10) },
    }),
    [chartData, series.length, theme]
  );

  return (
    <Card
      title="使用统计"
      loading={loading}
      extra={
        <Space wrap>
          <RangePicker
            value={range}
            onChange={(v) => v && setRange(v)}
            allowClear={false}
            disabledDate={(d) => d && d > dayjs().endOf('day')}
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={fetchStats}>
            查询
          </Button>
          <Button icon={<ReloadOutlined />} onClick={() => setRange(defaultRange())}>
            近 30 天
          </Button>
        </Space>
      }
    >
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={12} sm={8} md={4}>
          <Statistic title="单次查询" value={summary?.query_single ?? 0} />
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Statistic title="批量查询" value={summary?.query_batch ?? 0} />
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Statistic title="流访问" value={summary?.stream ?? 0} />
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Statistic title="查询成功率" value={pct(summary?.query_success_rate)} />
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Statistic title="流访问成功率" value={pct(summary?.stream_success_rate)} />
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Statistic title="活跃用户" value={summary?.active_users ?? 0} />
        </Col>
      </Row>

      <Card type="inner" title="每日调用趋势" style={{ marginBottom: 16 }}>
        {chartData.length > 0 ? (
          <Line {...lineConfig} />
        ) : (
          <Text type="secondary">所选时间范围内暂无数据</Text>
        )}
      </Card>

      <Text type="secondary">
        数据来自审计日志，默认保留 3 个月。
        {truncated ? ' 起始日期已截断至保留期内。' : ''}
        明细请见「审计查询」。
      </Text>
    </Card>
  );
}

export default Dashboard;
