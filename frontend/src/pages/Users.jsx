import { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  message,
  Popconfirm,
  Tag,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  KeyOutlined,
  DeleteOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { adminService } from '../services/authService';
import { useAuthStore } from '../store/authStore';
import { formatDateTime } from '../utils/format';

function Users() {
  const { user: currentUser } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState([]);

  // 新增用户
  const [createOpen, setCreateOpen] = useState(false);
  const [createForm] = Form.useForm();

  // 修改角色
  const [roleOpen, setRoleOpen] = useState(false);
  const [roleForm] = Form.useForm();
  const [roleTarget, setRoleTarget] = useState(null);

  // 重置密码
  const [pwdOpen, setPwdOpen] = useState(false);
  const [pwdForm] = Form.useForm();
  const [pwdTarget, setPwdTarget] = useState(null);

  const fetchList = async () => {
    setLoading(true);
    try {
      const res = await adminService.listUsers();
      if (res?.success) {
        setList(res.list || []);
      } else {
        message.error(res?.message || '获取用户列表失败');
      }
    } catch (err) {
      message.error(err?.response?.data?.message || '获取用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchList();
  }, []);

  const onCreate = async () => {
    try {
      const values = await createForm.validateFields();
      const res = await adminService.createUser(values);
      if (res?.success) {
        message.success('用户创建成功');
        setCreateOpen(false);
        createForm.resetFields();
        fetchList();
      } else {
        message.error(res?.message || '创建失败');
      }
    } catch (err) {
      if (err?.errorFields) return;
      message.error(err?.response?.data?.message || '创建失败');
    }
  };

  const openRole = (record) => {
    setRoleTarget(record);
    roleForm.setFieldsValue({ role: record.role });
    setRoleOpen(true);
  };

  const onUpdateRole = async () => {
    try {
      const values = await roleForm.validateFields();
      const res = await adminService.updateUserRole(roleTarget.id, values.role);
      if (res?.success) {
        message.success('角色已更新');
        setRoleOpen(false);
        fetchList();
      } else {
        message.error(res?.message || '更新失败');
      }
    } catch (err) {
      if (err?.errorFields) return;
      message.error(err?.response?.data?.message || '更新失败');
    }
  };

  const openResetPassword = (record) => {
    setPwdTarget(record);
    pwdForm.resetFields();
    setPwdOpen(true);
  };

  const onResetPassword = async () => {
    try {
      const values = await pwdForm.validateFields();
      const res = await adminService.resetUserPassword(pwdTarget.id, values.new_password);
      if (res?.success) {
        message.success('密码已重置');
        setPwdOpen(false);
      } else {
        message.error(res?.message || '重置失败');
      }
    } catch (err) {
      if (err?.errorFields) return;
      message.error(err?.response?.data?.message || '重置失败');
    }
  };

  const onDelete = async (record) => {
    try {
      const res = await adminService.deleteUser(record.id);
      if (res?.success) {
        message.success('用户已删除');
        fetchList();
      } else {
        message.error(res?.message || '删除失败');
      }
    } catch (err) {
      message.error(err?.response?.data?.message || '删除失败');
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
    { title: '用户名', dataIndex: 'username', key: 'username' },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      width: 100,
      render: (role) =>
        role === 'admin' ? <Tag color="purple">管理员</Tag> : <Tag>普通用户</Tag>,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: formatDateTime,
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 180,
      render: formatDateTime,
    },
    {
      title: '操作',
      key: 'action',
      width: 280,
      render: (_, record) => {
        const isSelf = record.username === currentUser?.username;
        return (
          <Space wrap>
            <Button
              size="small"
              icon={<EditOutlined />}
              onClick={() => openRole(record)}
              disabled={isSelf}
            >
              改角色
            </Button>
            <Button
              size="small"
              icon={<KeyOutlined />}
              onClick={() => openResetPassword(record)}
            >
              重置密码
            </Button>
            <Popconfirm
              title="确认删除该用户？"
              okText="删除"
              cancelText="取消"
              okButtonProps={{ danger: true }}
              onConfirm={() => onDelete(record)}
              disabled={isSelf}
            >
              <Button size="small" danger icon={<DeleteOutlined />} disabled={isSelf}>
                删除
              </Button>
            </Popconfirm>
          </Space>
        );
      },
    },
  ];

  return (
    <Card
      title="用户管理"
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchList}>
            刷新
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
            新增用户
          </Button>
        </Space>
      }
    >
      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={list}
        pagination={false}
      />

      <Modal
        title="新增用户"
        open={createOpen}
        onOk={onCreate}
        onCancel={() => setCreateOpen(false)}
        okText="创建"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={createForm} layout="vertical" initialValues={{ role: 'user' }}>
          <Form.Item
            name="username"
            label="用户名"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input autoComplete="off" />
          </Form.Item>
          <Form.Item
            name="password"
            label="密码"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码长度至少 6 位' },
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'user', label: '普通用户' },
                { value: 'admin', label: '管理员' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`修改角色 - ${roleTarget?.username || ''}`}
        open={roleOpen}
        onOk={onUpdateRole}
        onCancel={() => setRoleOpen(false)}
        okText="确定"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={roleForm} layout="vertical">
          <Form.Item name="role" label="角色" rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'user', label: '普通用户' },
                { value: 'admin', label: '管理员' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`重置密码 - ${pwdTarget?.username || ''}`}
        open={pwdOpen}
        onOk={onResetPassword}
        onCancel={() => setPwdOpen(false)}
        okText="重置"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={pwdForm} layout="vertical">
          <Form.Item
            name="new_password"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码长度至少 6 位' },
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}

export default Users;
