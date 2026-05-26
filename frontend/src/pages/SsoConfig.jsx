import { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Switch,
  Tag,
  message,
  Popconfirm,
  Typography,
  Alert,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { adminService } from '../services/authService';

const { Text } = Typography;

const DEFAULT_OIDC_CONFIG = {
  issuer: '',
  client_id: '',
  client_secret: '',
  redirect_url: '',
  scopes: ['openid', 'profile', 'email'],
  username_claim: 'preferred_username',
  skip_tls_verify: false,
};

function SsoConfig() {
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState([]);

  const [open, setOpen] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [form] = Form.useForm();

  const fetchList = async () => {
    setLoading(true);
    try {
      const res = await adminService.listSSOProvidersAdmin();
      if (res?.success) {
        setList(res.list || []);
      } else {
        message.error(res?.message || '加载失败');
      }
    } catch (err) {
      message.error(err?.response?.data?.message || '加载失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchList();
  }, []);

  const onAdd = () => {
    setEditingId(null);
    form.resetFields();
    form.setFieldsValue({
      name: 'OIDC',
      enabled: true,
      ...flatten(DEFAULT_OIDC_CONFIG),
    });
    setOpen(true);
  };

  const onEdit = (record) => {
    setEditingId(record.id);
    form.resetFields();
    form.setFieldsValue({
      name: record.name,
      enabled: record.enabled,
      ...flatten({ ...DEFAULT_OIDC_CONFIG, ...(record.config || {}) }),
    });
    setOpen(true);
  };

  const onToggle = async (record) => {
    try {
      const res = await adminService.toggleSSOProvider(record.id);
      if (res?.success) {
        message.success('状态已更新');
        fetchList();
      } else {
        message.error(res?.message || '更新失败');
      }
    } catch (err) {
      message.error(err?.response?.data?.message || '更新失败');
    }
  };

  const onDelete = async (record) => {
    try {
      const res = await adminService.deleteSSOProvider(record.id);
      if (res?.success) {
        message.success('已删除');
        fetchList();
      } else {
        message.error(res?.message || '删除失败');
      }
    } catch (err) {
      message.error(err?.response?.data?.message || '删除失败');
    }
  };

  const onSave = async () => {
    try {
      const values = await form.validateFields();
      const payload = {
        type: 'oidc',
        name: values.name,
        enabled: !!values.enabled,
        config: unflatten(values),
      };
      let res;
      if (editingId) {
        res = await adminService.updateSSOProvider(editingId, payload);
      } else {
        res = await adminService.createSSOProvider(payload);
      }
      if (res?.success) {
        message.success(editingId ? '已更新' : '已创建');
        setOpen(false);
        fetchList();
      } else {
        message.error(res?.message || '保存失败');
      }
    } catch (err) {
      if (err?.errorFields) return;
      message.error(err?.response?.data?.message || '保存失败');
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    {
      title: '类型',
      dataIndex: 'type',
      width: 90,
      render: () => <Tag color="blue">OIDC</Tag>,
    },
    { title: '名称', dataIndex: 'name' },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 100,
      render: (v, record) => (
        <Switch checked={v} onChange={() => onToggle(record)} checkedChildren="启用" unCheckedChildren="停用" />
      ),
    },
    {
      title: '回调地址',
      key: 'callback',
      render: (_, record) => (
        <Text copyable code style={{ fontSize: 12 }}>{`/api/auth/sso/oidc/${record.id}/callback`}</Text>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => onEdit(record)}>编辑</Button>
          <Popconfirm
            title="确认删除该提供商？"
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
            onConfirm={() => onDelete(record)}
          >
            <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="SSO 单点登录配置（OIDC）"
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchList}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={onAdd}>新增 OIDC</Button>
        </Space>
      }
    >
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message="说明"
        description={
          <>
            <div>• 仅支持 OpenID Connect (OIDC) 协议。</div>
            <div>• OIDC 回调地址：<code>/api/auth/sso/oidc/&lt;id&gt;/callback</code>，需要在 IdP 端登记。</div>
            <div>• SSO 登录的用户会自动以「普通用户」身份创建，可在「用户管理」里手动改为管理员。</div>
          </>
        }
      />

      <Table rowKey="id" loading={loading} columns={columns} dataSource={list} pagination={false} />

      <Modal
        title={editingId ? '编辑 OIDC 提供商' : '新增 OIDC 提供商'}
        open={open}
        onOk={onSave}
        onCancel={() => setOpen(false)}
        okText="保存"
        cancelText="取消"
        width={700}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="显示名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="例如：公司 Keycloak" />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="issuer" label="Issuer URL" rules={[{ required: true }]}>
            <Input placeholder="https://accounts.google.com" />
          </Form.Item>
          <Form.Item name="client_id" label="Client ID" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="client_secret" label="Client Secret" rules={[{ required: true }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item
            name="redirect_url"
            label="Redirect URL"
            rules={[{ required: true }]}
            tooltip="必须等于 IdP 中登记的回调地址，通常形如 https://your.host/api/auth/sso/oidc/<id>/callback"
          >
            <Input placeholder="https://your.host/api/auth/sso/oidc/1/callback" />
          </Form.Item>
          <Form.Item name="scopes_str" label="Scopes" tooltip="逗号分隔；默认 openid,profile,email">
            <Input placeholder="openid,profile,email" />
          </Form.Item>
          <Form.Item
            name="username_claim"
            label="用户名 Claim"
            tooltip="默认 preferred_username，缺失会回退到 email、sub"
          >
            <Input placeholder="preferred_username" />
          </Form.Item>
          <Form.Item name="skip_tls_verify" label="跳过 TLS 校验" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}

function flatten(cfg) {
  const o = { ...cfg };
  if (Array.isArray(o.scopes)) {
    o.scopes_str = o.scopes.join(',');
    delete o.scopes;
  }
  return o;
}

function unflatten(values) {
  const scopes = (values.scopes_str || 'openid,profile,email')
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
  return {
    issuer: values.issuer || '',
    client_id: values.client_id || '',
    client_secret: values.client_secret || '',
    redirect_url: values.redirect_url || '',
    scopes,
    username_claim: values.username_claim || 'preferred_username',
    skip_tls_verify: !!values.skip_tls_verify,
  };
}

export default SsoConfig;
