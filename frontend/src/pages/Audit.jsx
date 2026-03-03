import { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  message,
  Form,
  Input,
  Select,
  DatePicker,
  Tag,
} from 'antd';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import { adminService } from '../services/authService';

const { RangePicker } = DatePicker;

function formatDateTime(v) {
  if (!v) return '-';
  const d = new Date(v);
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

const ACTION_OPTIONS = [
  { value: '', label: '全部' },
  { value: 'login_success', label: '登录成功' },
  { value: 'login_fail', label: '登录失败' },
  { value: 'play', label: '播放' },
  { value: 'play_batch', label: '批量播放' },
  { value: 'config_save', label: '保存配置' },
  { value: 'config_reload', label: '重新加载配置' },
];

function Audit() {
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [form] = Form.useForm();

  const fetchLogs = useCallback(
    async (pageNum = 1, pageSizeNum = pageSize) => {
      setLoading(true);
      try {
        const values = form.getFieldsValue();
        const params = {
          page: pageNum,
          page_size: pageSizeNum,
        };
        if (values.action) params.action = values.action;
        if (values.username?.trim()) params.username = values.username.trim();
        if (values.range?.length === 2) {
          const start = new Date(values.range[0]);
          const end = new Date(values.range[1]);
          end.setHours(23, 59, 59, 999);
          params.from = start.toISOString();
          params.to = end.toISOString();
        }
        const res = await adminService.getAuditLogs(params);
        if (res?.success) {
          setList(res.list || []);
          setTotal(res.total ?? 0);
          setPage(res.page ?? pageNum);
          setPageSize(res.page_size ?? pageSizeNum);
        } else {
          message.error(res?.message || '获取审计日志失败');
        }
      } catch (err) {
        message.error(err?.response?.data?.message || '获取审计日志失败');
      } finally {
        setLoading(false);
      }
    },
    [form, pageSize]
  );

  useEffect(() => {
    fetchLogs(1, pageSize);
  }, []);

  const onSearch = () => {
    setPage(1);
    fetchLogs(1, pageSize);
  };

  const onReset = () => {
    form.resetFields();
    setPage(1);
    fetchLogs(1, pageSize);
  };

  const onTableChange = (pagination) => {
    setPage(pagination.current);
    setPageSize(pagination.pageSize);
    fetchLogs(pagination.current, pagination.pageSize);
  };

  const columns = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (v) => formatDateTime(v),
    },
    {
      title: '动作',
      dataIndex: 'action',
      key: 'action',
      width: 120,
      render: (action) => {
        const opt = ACTION_OPTIONS.find((o) => o.value === action);
        return opt ? opt.label : action || '-';
      },
    },
    { title: '用户', dataIndex: 'username', key: 'username', width: 100, ellipsis: true },
    { title: '角色', dataIndex: 'role', key: 'role', width: 80 },
    { title: '客户端 IP', dataIndex: 'client_ip', key: 'client_ip', width: 120 },
    {
      title: '资源',
      dataIndex: 'resource',
      key: 'resource',
      minWidth: 220,
      render: (v) =>
        v ? <span style={{ wordBreak: 'break-all', display: 'block' }}>{v}</span> : '-',
    },
    { title: '详情', dataIndex: 'detail', key: 'detail', ellipsis: true },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status) =>
        status === 'success' ? (
          <Tag color="success">成功</Tag>
        ) : (
          <Tag color="error">失败</Tag>
        ),
    },
  ];

  return (
    <Card title="审计查询">
      <Form form={form} layout="inline" style={{ marginBottom: 16 }} onFinish={onSearch}>
        <Form.Item name="range" label="时间范围">
          <RangePicker showTime />
        </Form.Item>
        <Form.Item name="action" label="动作类型">
          <Select options={ACTION_OPTIONS} style={{ width: 140 }} allowClear />
        </Form.Item>
        <Form.Item name="username" label="用户名">
          <Input placeholder="用户名" allowClear style={{ width: 120 }} />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" icon={<SearchOutlined />} loading={loading}>
              查询
            </Button>
            <Button icon={<ReloadOutlined />} onClick={onReset}>
              重置
            </Button>
          </Space>
        </Form.Item>
      </Form>
      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={list}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (t) => `共 ${t} 条`,
          pageSizeOptions: ['20', '50', '100'],
        }}
        onChange={onTableChange}
        scroll={{ x: 1000 }}
      />
    </Card>
  );
}

export default Audit;
