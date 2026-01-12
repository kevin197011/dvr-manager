import { useState } from 'react';
import {
  Card,
  Input,
  Button,
  Table,
  message,
  Space,
  Typography,
  Tag,
  Spin,
} from 'antd';
import { SearchOutlined, PlayCircleOutlined, DownloadOutlined } from '@ant-design/icons';
import { dvrService } from '../services/authService';
import VideoPlayer from '../components/VideoPlayer';

const { Title, Text } = Typography;
const { TextArea } = Input;

function Home() {
  const [loading, setLoading] = useState(false);
  const [recordIds, setRecordIds] = useState('');
  const [results, setResults] = useState([]);

  const handleQuery = async () => {
    if (!recordIds.trim()) {
      message.warning('请输入录像编号');
      return;
    }

    const ids = recordIds
      .split('\n')
      .map((id) => id.trim())
      .filter((id) => id);

    if (ids.length === 0) {
      message.warning('请输入有效的录像编号');
      return;
    }

    setLoading(true);
    try {
      let response;
      if (ids.length === 1) {
        response = await dvrService.play(ids[0]);
        if (response.success) {
          setResults([
            {
              key: ids[0],
              recordId: ids[0],
              found: true,
              proxyUrl: response.proxy_url,
              playing: false,
            },
          ]);
        } else {
          setResults([
            {
              key: ids[0],
              recordId: ids[0],
              found: false,
              playing: false,
            },
          ]);
        }
      } else {
        response = await dvrService.batchPlay(ids);
        if (response.success && response.results) {
          setResults(
            response.results.map((r, index) => {
              // 确保 recordId 字段存在，兼容不同的字段名
              const recordId = r.record_id || r.recordId || ids[index] || `record-${index}`;
              return {
                key: recordId || `key-${index}`,
                recordId: recordId,
                found: r.found !== undefined ? r.found : false,
                proxyUrl: r.proxy_url || r.proxyUrl || null,
                playing: false,
              };
            })
          );
        }
      }
    } catch (error) {
      message.error('查询失败：' + (error.message || '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  const handleTogglePlay = (recordId) => {
    setResults(results.map((r) => {
      if (r.recordId === recordId) {
        return { ...r, playing: !r.playing };
      }
      // 关闭其他正在播放的视频
      return { ...r, playing: false };
    }));
  };

  const handleDownload = async (recordId, proxyUrl) => {
    try {
      const hide = message.loading({ 
        content: `正在下载 ${recordId}.mp4...`, 
        key: `download-${recordId}`,
        duration: 0 
      });
      
      // 使用 fetch 下载视频
      const response = await fetch(proxyUrl);
      if (!response.ok) {
        hide();
        throw new Error(`下载失败: ${response.status} ${response.statusText}`);
      }

      // 创建 Blob
      const blob = await response.blob();
      
      // 创建下载链接
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${recordId}.mp4`;
      document.body.appendChild(a);
      a.click();
      
      // 清理
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
      hide();
      message.success({ 
        content: `${recordId}.mp4 下载完成`, 
        key: `download-${recordId}`,
        duration: 3 
      });
    } catch (error) {
      message.error({ 
        content: `下载失败：${error.message || '未知错误'}`, 
        key: `download-${recordId}`,
        duration: 5 
      });
    }
  };

  const columns = [
    {
      title: '录像编号',
      dataIndex: 'recordId',
      key: 'recordId',
      render: (text, record) => {
        // 确保显示录像编号，兼容多种字段名
        return text || record?.recordId || record?.record_id || '-';
      },
    },
    {
      title: '状态',
      dataIndex: 'found',
      key: 'found',
      render: (found) => (
        <Tag color={found ? 'success' : 'error'}>
          {found ? '已找到' : '未找到'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space>
          {record.found && (
            <>
              <Button
                type="link"
                icon={<PlayCircleOutlined />}
                onClick={() => handleTogglePlay(record.recordId)}
              >
                {record.playing ? '关闭' : '播放'}
              </Button>
              <Button
                type="link"
                icon={<DownloadOutlined />}
                onClick={() => handleDownload(record.recordId, record.proxyUrl)}
              >
                下载
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Title level={2}>录像查询</Title>
      <Card>
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <div>
            <Text strong>请输入录像编号（每行一个，支持批量查询）：</Text>
            <TextArea
              rows={6}
              value={recordIds}
              onChange={(e) => setRecordIds(e.target.value)}
              placeholder="例如：GT03225A120DV"
              style={{ marginTop: 8 }}
            />
          </div>
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={handleQuery}
            loading={loading}
            size="large"
          >
            查询
          </Button>
        </Space>
      </Card>

      {results.length > 0 && (
        <Card title="查询结果" style={{ marginTop: 24 }}>
          <Table
            columns={columns}
            dataSource={results}
            pagination={false}
            loading={loading}
            expandable={{
              expandedRowKeys: results.filter(r => r.playing).map(r => r.key),
              onExpandedRowsChange: (expandedKeys) => {
                setResults(results.map(r => ({
                  ...r,
                  playing: expandedKeys.includes(r.key),
                })));
              },
              expandedRowRender: (record) => {
                if (!record.found || !record.proxyUrl) {
                  return null;
                }
                return (
                  <div style={{ padding: '16px 0' }}>
                    <VideoPlayer 
                      key={record.proxyUrl} 
                      src={record.proxyUrl} 
                      autoPlay={record.playing}
                    />
                  </div>
                );
              },
              rowExpandable: (record) => record.found && !!record.proxyUrl,
            }}
          />
        </Card>
      )}
    </div>
  );
}

export default Home;
