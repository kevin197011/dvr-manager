import { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Input,
  Space,
  message,
  Popconfirm,
  Typography,
  Form,
  InputNumber,
  Switch,
  Row,
  Col,
  Tabs,
  Descriptions,
  Divider,
  Alert,
  Tooltip,
  Tag,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  ReloadOutlined,
  SaveOutlined,
  CloudServerOutlined,
  SettingOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { adminService } from '../services/authService';

const { Title, Text } = Typography;

function Admin() {
  const [loading, setLoading] = useState(false);
  const [servers, setServers] = useState([]);
  const [config, setConfig] = useState(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [serversRes, configRes] = await Promise.all([
        adminService.getDVRServers(),
        adminService.getConfig(),
      ]);

      if (serversRes.success) {
        setServers(
          serversRes.servers.map((server, index) => ({
            key: index,
            server,
          }))
        );
      }

      if (configRes.success) {
        setConfig(configRes.config);
        // 转换时间单位：后端 time.Duration 序列化为纳秒（int64），前端显示秒
        const config = configRes.config;
        const formValues = {
          ...config,
          server: {
            ...config.server,
            timeout: config.server?.timeout 
              ? (typeof config.server.timeout === 'number' 
                  ? Math.floor(config.server.timeout / 1e9) 
                  : config.server.timeout)
              : 30, // 默认30秒
          },
          dvr: {
            ...config.dvr,
            timeout: config.dvr?.timeout 
              ? (typeof config.dvr.timeout === 'number' 
                  ? Math.floor(config.dvr.timeout / 1e9) 
                  : config.dvr.timeout)
              : 10, // 默认10秒
          },
        };
        form.setFieldsValue(formValues);
      }
    } catch (error) {
      message.error('加载数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleAddServer = () => {
    setServers([...servers, { key: Date.now(), server: '' }]);
  };

  const handleDeleteServer = (key) => {
    setServers(servers.filter((item) => item.key !== key));
  };

  const handleServerChange = (key, value) => {
    setServers(
      servers.map((item) =>
        item.key === key ? { ...item, server: value } : item
      )
    );
  };

  const handleSaveAll = async () => {
    // 收集所有配置
    const formValues = form.getFieldsValue();
    const serverList = servers
      .map((item) => item.server.trim())
      .filter((s) => s);

    if (serverList.length === 0) {
      message.warning('至少需要添加一个服务器');
      return;
    }

    setLoading(true);
    try {
      // 构建完整配置对象
      // 前端输入的是秒，后端需要的是秒数（会在后端转换为 time.Duration）
      const configData = {
        server: {
          ...formValues.server,
          timeout: formValues.server?.timeout || 30, // 默认30秒
        },
        dvr: {
          ...formValues.dvr,
          timeout: formValues.dvr?.timeout || 10, // 默认10秒
        },
        dvr_servers: serverList,
        cors: formValues.cors || {},
      };

      const response = await adminService.updateConfig(configData);
      if (response.success) {
        message.success('配置保存成功！');
        loadData();
      } else {
        message.error('保存失败：' + (response.message || '未知错误'));
      }
    } catch (error) {
      message.error('保存失败：' + (error.message || '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  const handleReload = async () => {
    setLoading(true);
    try {
      const response = await adminService.reloadConfig();
      if (response.success) {
        message.success('配置重新加载成功');
        loadData();
      }
    } catch (error) {
      message.error('重新加载失败');
    } finally {
      setLoading(false);
    }
  };

  const serverColumns = [
    {
      title: '序号',
      key: 'index',
      width: 80,
      render: (_, __, index) => (
        <Text strong style={{ color: '#1890ff' }}>#{index + 1}</Text>
      ),
    },
    {
      title: '服务器地址',
      dataIndex: 'server',
      key: 'server',
      render: (text, record) => (
        <Input
          value={text}
          onChange={(e) => handleServerChange(record.key, e.target.value)}
          placeholder="http://example.com:8080/record"
          size="large"
        />
      ),
    },
    {
      title: '状态',
      key: 'status',
      width: 100,
      render: (_, record) => (
        record.server && record.server.trim() ? (
          <Tag icon={<CheckCircleOutlined />} color="success">已配置</Tag>
        ) : (
          <Tag color="default">未配置</Tag>
        )
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_, record) => (
        <Popconfirm
          title="确定要删除这个服务器吗？"
          description="删除后需要保存配置才能生效"
          onConfirm={() => handleDeleteServer(record.key)}
          okText="确定"
          cancelText="取消"
        >
          <Button type="link" danger icon={<DeleteOutlined />}>
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <Title level={2} style={{ margin: 0 }}>系统管理</Title>
          <Text type="secondary" style={{ fontSize: 14 }}>
            配置 DVR 服务器和系统参数
          </Text>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={loadData}>
            刷新
          </Button>
          <Button
            type="primary"
            icon={<SaveOutlined />}
            onClick={handleSaveAll}
            loading={loading}
            size="large"
          >
            保存所有配置
          </Button>
        </Space>
      </div>

      <Tabs 
        defaultActiveKey="servers" 
        size="large"
        items={[
          {
            key: 'servers',
            label: (
              <span>
                <CloudServerOutlined />
                DVR 服务器
              </span>
            ),
            children: (
              <Card>
                <Alert
                  message="DVR 服务器配置"
                  description={
                    <div>
                      <Text>配置用于查询录像的 DVR 服务器地址列表。系统会按顺序查询这些服务器，直到找到录像为止。</Text>
                      <br />
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        格式示例：http://dvr.example.com:8080/record
                      </Text>
                    </div>
                  }
                  type="info"
                  icon={<InfoCircleOutlined />}
                  showIcon
                  style={{ marginBottom: 24 }}
                />
                
                <Space direction="vertical" style={{ width: '100%' }} size="middle">
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Text strong>已配置 {servers.length} 个服务器</Text>
                    <Button
                      icon={<PlusOutlined />}
                      onClick={handleAddServer}
                      type="primary"
                      ghost
                    >
                      添加服务器
                    </Button>
                  </div>
                  
                  {servers.length > 0 ? (
                    <Table
                      columns={serverColumns}
                      dataSource={servers}
                      pagination={false}
                      loading={loading}
                      size="middle"
                    />
                  ) : (
                    <Card style={{ textAlign: 'center', padding: '40px 0' }}>
                      <Text type="secondary">暂无服务器配置，请点击上方"添加服务器"按钮添加</Text>
                    </Card>
                  )}
                </Space>
              </Card>
            ),
          },
          {
            key: 'settings',
            label: (
              <span>
                <SettingOutlined />
                系统参数
              </span>
            ),
            children: (
              <Card>
                <Alert
                  message="系统参数配置"
                  description="配置服务器运行参数和 DVR 查询行为"
                  type="info"
                  icon={<InfoCircleOutlined />}
                  showIcon
                  style={{ marginBottom: 24 }}
                />

                <Form
                  form={form}
                  layout="vertical"
                  initialValues={config}
                >
                  <Title level={4} style={{ marginTop: 0 }}>
                    <SettingOutlined style={{ marginRight: 8 }} />
                    服务器配置
                  </Title>
                  <Row gutter={[24, 16]}>
                    <Col span={12}>
                      <Form.Item 
                        label={
                          <span>
                            服务器端口
                            <Tooltip title="后端服务监听的端口号">
                              <InfoCircleOutlined style={{ marginLeft: 4, color: '#999' }} />
                            </Tooltip>
                          </span>
                        }
                        name={['server', 'port']}
                      >
                        <InputNumber 
                          min={1} 
                          max={65535} 
                          style={{ width: '100%' }}
                          placeholder="默认: 8080"
                        />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item 
                        label={
                          <span>
                            请求超时（秒）
                            <Tooltip title="HTTP 请求的总超时时间">
                              <InfoCircleOutlined style={{ marginLeft: 4, color: '#999' }} />
                            </Tooltip>
                          </span>
                        }
                        name={['server', 'timeout']}
                        initialValue={30}
                      >
                        <InputNumber 
                          min={1} 
                          style={{ width: '100%' }}
                          placeholder="默认: 30"
                          addonAfter="秒"
                        />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Divider />

                  <Title level={4}>
                    <CloudServerOutlined style={{ marginRight: 8 }} />
                    DVR 查询配置
                  </Title>
                  <Row gutter={[24, 16]}>
                    <Col span={12}>
                      <Form.Item 
                        label={
                          <span>
                            DVR 查询超时（秒）
                            <Tooltip title="单个 DVR 服务器查询的超时时间，建议 10-15 秒">
                              <InfoCircleOutlined style={{ marginLeft: 4, color: '#999' }} />
                            </Tooltip>
                          </span>
                        }
                        name={['dvr', 'timeout']}
                        initialValue={10}
                      >
                        <InputNumber 
                          min={1} 
                          style={{ width: '100%' }}
                          placeholder="默认: 10"
                          addonAfter="秒"
                        />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item 
                        label={
                          <span>
                            重试次数
                            <Tooltip title="查询失败时的重试次数，建议 2-3 次">
                              <InfoCircleOutlined style={{ marginLeft: 4, color: '#999' }} />
                            </Tooltip>
                          </span>
                        }
                        name={['dvr', 'retry']}
                      >
                        <InputNumber 
                          min={0} 
                          max={10} 
                          style={{ width: '100%' }}
                          placeholder="默认: 3"
                          addonAfter="次"
                        />
                      </Form.Item>
                    </Col>
                  </Row>
                  <Row gutter={[24, 16]}>
                    <Col span={24}>
                      <Form.Item
                        label={
                          <span>
                            跳过 TLS 证书验证
                            <Tooltip title="适用于自签名证书或内网环境，生产环境建议关闭">
                              <InfoCircleOutlined style={{ marginLeft: 4, color: '#999' }} />
                            </Tooltip>
                          </span>
                        }
                        name={['dvr', 'skip_tls_verify']}
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>
                    </Col>
                  </Row>
                </Form>
              </Card>
            ),
          },
        ]}
      />
    </div>
  );
}

export default Admin;
