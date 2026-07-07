import { useRef, useState } from 'react';
import {
  Card,
  Input,
  Button,
  Table,
  message,
  Space,
  Typography,
  Tag,
  Tooltip,
} from 'antd';
import { SearchOutlined, PlayCircleOutlined, DownloadOutlined } from '@ant-design/icons';
import { dvrService } from '../services/authService';
import { getApiErrorMessage } from '../utils/format';
import VideoPlayer from '../components/VideoPlayer';

const { Title, Text } = Typography;
const { TextArea } = Input;

function mapPlayError(error) {
  if (error?.name === 'CanceledError' || error?.code === 'ERR_CANCELED') {
    return null;
  }
  if (error?.response?.status === 404) {
    return '未找到';
  }
  return getApiErrorMessage(error, '查询失败');
}

function Home() {
  const [loading, setLoading] = useState(false);
  const [recordIds, setRecordIds] = useState('');
  const [results, setResults] = useState([]);
  const queryAbortRef = useRef(null);

  const handleQuery = async () => {
    if (!recordIds.trim()) {
      message.warning('请输入录像编号');
      return;
    }

    const ids = recordIds
      .split('\n')
      .map((id) => id.trim())
      .filter(Boolean);

    if (ids.length === 0) {
      message.warning('请输入有效的录像编号');
      return;
    }

    queryAbortRef.current?.abort();
    const controller = new AbortController();
    queryAbortRef.current = controller;

    setLoading(true);
    try {
      if (ids.length === 1) {
        const response = await dvrService.play(ids[0], { signal: controller.signal });
        if (response?.success) {
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
              error: response?.message || '未找到',
            },
          ]);
        }
      } else {
        const response = await dvrService.batchPlay(ids, { signal: controller.signal });
        if (response?.success && response.results) {
          setResults(
            response.results.map((r, index) => {
              const recordId = r.record_id || r.recordId || ids[index] || `record-${index}`;
              return {
                key: recordId || `key-${index}`,
                recordId,
                found: !!r.found,
                proxyUrl: r.proxy_url || r.proxyUrl || null,
                playing: false,
                error: r.found ? undefined : r.message || '未找到',
              };
            })
          );
        } else {
          const err = response?.message || '查询失败';
          setResults(
            ids.map((id, index) => ({
              key: id || `key-${index}`,
              recordId: id,
              found: false,
              playing: false,
              error: err,
            }))
          );
        }
      }
    } catch (error) {
      const errorMsg = mapPlayError(error);
      if (errorMsg === null) {
        return;
      }
      setResults(
        ids.map((id, index) => ({
          key: id || `key-${index}`,
          recordId: id,
          found: false,
          playing: false,
          error: errorMsg,
        }))
      );
    } finally {
      if (queryAbortRef.current === controller) {
        setLoading(false);
      }
    }
  };

  const handleTogglePlay = (recordId) => {
    setResults((prev) =>
      prev.map((r) => {
        if (r.recordId === recordId) {
          return { ...r, playing: !r.playing };
        }
        return { ...r, playing: false };
      })
    );
  };

  const handleDownload = (recordId, proxyUrl) => {
    const a = document.createElement('a');
    a.href = proxyUrl;
    a.download = `${recordId}.mp4`;
    a.rel = 'noopener';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    message.success(`已开始下载 ${recordId}.mp4`);
  };

  const columns = [
    {
      title: '录像编号',
      dataIndex: 'recordId',
      key: 'recordId',
      render: (text, record) => text || record?.recordId || '-',
    },
    {
      title: '状态',
      dataIndex: 'found',
      key: 'found',
      render: (found, record) => {
        if (found) {
          return <Tag color="success">已找到</Tag>;
        }
        const errorTag = <Tag color="error">未找到</Tag>;
        return record.error ? <Tooltip title={record.error}>{errorTag}</Tooltip> : errorTag;
      },
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
              expandedRowKeys: results.filter((r) => r.playing).map((r) => r.key),
              onExpandedRowsChange: (expandedKeys) => {
                setResults((prev) =>
                  prev.map((r) => ({
                    ...r,
                    playing: expandedKeys.includes(r.key),
                  }))
                );
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
