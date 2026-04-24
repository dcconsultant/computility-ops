import { useEffect, useMemo, useState } from 'react';
import { Button, Card, DatePicker, Form, Input, InputNumber, Popconfirm, Space, Table, Tag, Upload, message } from 'antd';
import { DeleteOutlined, DownloadOutlined, UploadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import type { UploadProps } from 'antd';
import { createContract, deleteContract, deleteContractAttachment, downloadContractAttachment, listContracts, updateContract, uploadContractAttachment } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { Contract } from '../types';

interface ContractFormValue {
  contract_name: string;
  period_start: dayjs.Dayjs;
  period_end: dayjs.Dayjs;
  pre_tax_amount: number;
  supplier: string;
  business_contact: string;
  tech_contact: string;
}

export default function ContractPage() {
  const [form] = Form.useForm<ContractFormValue>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [contracts, setContracts] = useState<Contract[]>([]);
  const [editingID, setEditingID] = useState<string>('');
  const [uploadingID, setUploadingID] = useState<string>('');

  async function reload() {
    setLoading(true);
    try {
      const resp = ensureApiOk(await listContracts());
      setContracts(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载合同列表失败'));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    reload();
  }, []);

  const canSubmit = useMemo(() => !saving && !uploadingID, [saving, uploadingID]);

  function resetForm() {
    setEditingID('');
    form.resetFields();
  }

  async function onSubmit(values: ContractFormValue) {
    setSaving(true);
    try {
      const payload = {
        contract_name: values.contract_name.trim(),
        period_start: values.period_start.format('YYYY-MM-DD'),
        period_end: values.period_end.format('YYYY-MM-DD'),
        pre_tax_amount: Number(values.pre_tax_amount || 0),
        supplier: values.supplier.trim(),
        business_contact: values.business_contact.trim(),
        tech_contact: values.tech_contact.trim()
      };
      if (editingID) {
        ensureApiOk(await updateContract(editingID, payload));
        message.success('合同已更新');
      } else {
        ensureApiOk(await createContract(payload));
        message.success('合同已创建');
      }
      resetForm();
      await reload();
    } catch (e) {
      message.error(parseApiError(e, editingID ? '更新合同失败' : '创建合同失败'));
    } finally {
      setSaving(false);
    }
  }

  function onEdit(item: Contract) {
    setEditingID(item.contract_id);
    form.setFieldsValue({
      contract_name: item.contract_name,
      period_start: dayjs(item.period_start),
      period_end: dayjs(item.period_end),
      pre_tax_amount: Number(item.pre_tax_amount || 0),
      supplier: item.supplier,
      business_contact: item.business_contact,
      tech_contact: item.tech_contact
    });
  }

  async function onDelete(contractId: string) {
    try {
      ensureApiOk(await deleteContract(contractId));
      message.success('合同已删除');
      if (editingID === contractId) {
        resetForm();
      }
      await reload();
    } catch (e) {
      message.error(parseApiError(e, '删除合同失败'));
    }
  }

  async function onDeleteAttachment(contractId: string, attachmentId: string) {
    try {
      ensureApiOk(await deleteContractAttachment(contractId, attachmentId));
      message.success('附件已删除');
      await reload();
    } catch (e) {
      message.error(parseApiError(e, '删除附件失败'));
    }
  }

  function uploadProps(contractId: string): UploadProps {
    return {
      maxCount: 1,
      showUploadList: false,
      customRequest: async (options) => {
        const file = options.file as File;
        setUploadingID(contractId);
        try {
          ensureApiOk(await uploadContractAttachment(contractId, file));
          message.success('附件上传成功');
          await reload();
          options.onSuccess?.({}, new XMLHttpRequest());
        } catch (e) {
          message.error(parseApiError(e, '附件上传失败'));
          options.onError?.(new Error('upload failed'));
        } finally {
          setUploadingID('');
        }
      }
    };
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card title="合同管理 - 新建/编辑合同" extra={editingID ? <Tag color="blue">编辑中：{editingID}</Tag> : <Tag>新建</Tag>}>
        <Form
          form={form}
          layout="vertical"
          onFinish={onSubmit}
          initialValues={{ pre_tax_amount: 0 }}
        >
          <Space wrap style={{ width: '100%' }}>
            <Form.Item name="contract_name" label="合同名称" rules={[{ required: true, message: '请输入合同名称' }]} style={{ width: 320 }}>
              <Input placeholder="例如：2026 年云资源采购框架合同" />
            </Form.Item>
            <Form.Item name="pre_tax_amount" label="合同税前金额" rules={[{ required: true, message: '请输入税前金额' }]} style={{ width: 260 }}>
              <InputNumber<number>
                min={0}
                precision={2}
                step={1000}
                style={{ width: '100%' }}
                formatter={(v) => Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                parser={(v) => Number((v || '').replace(/,/g, ''))}
              />
            </Form.Item>
            <Form.Item name="supplier" label="合同供应商" rules={[{ required: true, message: '请输入供应商' }]} style={{ width: 260 }}>
              <Input placeholder="例如：XX 科技有限公司" />
            </Form.Item>
          </Space>

          <Space wrap style={{ width: '100%' }}>
            <Form.Item name="period_start" label="合同开始日期" rules={[{ required: true, message: '请选择开始日期' }]}>
              <DatePicker />
            </Form.Item>
            <Form.Item name="period_end" label="合同结束日期" rules={[{ required: true, message: '请选择结束日期' }]}>
              <DatePicker />
            </Form.Item>
            <Form.Item name="business_contact" label="商务接口人" rules={[{ required: true, message: '请输入商务接口人' }]} style={{ width: 220 }}>
              <Input placeholder="姓名/邮箱/电话" />
            </Form.Item>
            <Form.Item name="tech_contact" label="技术接口人" rules={[{ required: true, message: '请输入技术接口人' }]} style={{ width: 220 }}>
              <Input placeholder="姓名/邮箱/电话" />
            </Form.Item>
          </Space>

          <Space>
            <Button type="primary" htmlType="submit" loading={saving} disabled={!canSubmit}>{editingID ? '保存更新' : '创建合同'}</Button>
            <Button onClick={resetForm} disabled={saving}>重置</Button>
          </Space>
        </Form>
      </Card>

      <Card title="合同管理 - 合同列表" extra={<Button onClick={reload} loading={loading}>刷新</Button>}>
        <Table
          rowKey="contract_id"
          loading={loading}
          dataSource={contracts}
          pagination={{ pageSize: 10, showTotal: (total) => `共 ${total} 条` }}
          scroll={{ x: 1600 }}
          columns={[
            { title: '合同ID', dataIndex: 'contract_id', width: 180 },
            { title: '合同名称', dataIndex: 'contract_name', width: 220 },
            { title: '合同期间', width: 220, render: (_: unknown, r: Contract) => `${r.period_start} ~ ${r.period_end}` },
            {
              title: '税前金额',
              dataIndex: 'pre_tax_amount',
              width: 150,
              render: (v: number) => Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
            },
            { title: '供应商', dataIndex: 'supplier', width: 180 },
            { title: '商务接口人', dataIndex: 'business_contact', width: 150 },
            { title: '技术接口人', dataIndex: 'tech_contact', width: 150 },
            {
              title: '附件',
              width: 320,
              render: (_: unknown, r: Contract) => (
                <Space direction="vertical" size={6} style={{ width: '100%' }}>
                  {(r.attachments || []).length === 0 ? <span style={{ color: '#999' }}>暂无附件</span> : (r.attachments || []).map((att) => (
                    <Space key={att.attachment_id}>
                      <a onClick={() => downloadContractAttachment(r.contract_id, att.attachment_id)}>{att.file_name}</a>
                      <span style={{ color: '#888' }}>{formatFileSize(att.file_size)}</span>
                      <Button type="link" size="small" icon={<DownloadOutlined />} onClick={() => downloadContractAttachment(r.contract_id, att.attachment_id)}>
                        下载
                      </Button>
                      <Popconfirm title="确认删除该附件？" onConfirm={() => onDeleteAttachment(r.contract_id, att.attachment_id)}>
                        <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
                      </Popconfirm>
                    </Space>
                  ))}
                </Space>
              )
            },
            {
              title: '操作',
              fixed: 'right',
              width: 220,
              render: (_: unknown, r: Contract) => (
                <Space>
                  <Button size="small" onClick={() => onEdit(r)}>编辑</Button>
                  <Upload {...uploadProps(r.contract_id)}>
                    <Button size="small" loading={uploadingID === r.contract_id} icon={<UploadOutlined />}>上传附件</Button>
                  </Upload>
                  <Popconfirm title="确认删除该合同及全部附件？" onConfirm={() => onDelete(r.contract_id)}>
                    <Button size="small" danger>删除</Button>
                  </Popconfirm>
                </Space>
              )
            }
          ]}
        />
      </Card>
    </Space>
  );
}

function formatFileSize(size: number) {
  if (size >= 1024 * 1024) {
    return `${(size / 1024 / 1024).toFixed(2)} MB`;
  }
  if (size >= 1024) {
    return `${(size / 1024).toFixed(1)} KB`;
  }
  return `${size || 0} B`;
}
