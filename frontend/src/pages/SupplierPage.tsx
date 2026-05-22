import { useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Input, Popconfirm, Space, Table, Tag, Upload, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { UploadProps } from 'antd';
import { createSupplier, deleteSupplier, exportSupplierTemplate, exportSuppliers, importSuppliers, listSuppliers, updateSupplier } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { Supplier } from '../types';

interface SupplierFormValue {
  company_full_name: string;
  tax_number: string;
  project_owner: string;
  project_owner_phone: string;
  tech_contact: string;
  tech_contact_phone: string;
  business_scope: string;
}

const EMPTY_FORM: SupplierFormValue = {
  company_full_name: '',
  tax_number: '',
  project_owner: '',
  project_owner_phone: '',
  tech_contact: '',
  tech_contact_phone: '',
  business_scope: ''
};

const TAX_REGEX = /^[0-9A-Z]{15,20}$/;
const PHONE_REGEX = /^(1\d{10}|0\d{2,3}-?\d{7,8})$/;

export default function SupplierPage() {
  const [form] = Form.useForm<SupplierFormValue>();
  const [editingID, setEditingID] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [query, setQuery] = useState('');
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);

  const columns: ColumnsType<Supplier> = useMemo(() => [
    { title: '公司全名', dataIndex: 'company_full_name', width: 260 },
    { title: '税号', dataIndex: 'tax_number', width: 220 },
    { title: '项目负责人', dataIndex: 'project_owner', width: 120 },
    { title: '负责人电话', dataIndex: 'project_owner_phone', width: 140 },
    { title: '技术接口人', dataIndex: 'tech_contact', width: 120 },
    { title: '技术电话', dataIndex: 'tech_contact_phone', width: 140 },
    { title: '业务范围', dataIndex: 'business_scope', ellipsis: true },
    {
      title: '操作',
      key: 'actions',
      width: 180,
      fixed: 'right',
      render: (_, row) => (
        <Space>
          <Button size="small" onClick={() => onEdit(row)}>编辑</Button>
          <Popconfirm title="确认删除该供应商？若已被合同引用将无法删除。" onConfirm={() => onDelete(row.supplier_id)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      )
    }
  ], []);

  const uploadProps: UploadProps = {
    accept: '.xlsx',
    showUploadList: false,
    beforeUpload: async (file) => {
      try {
        const resp = ensureApiOk(await importSuppliers(file));
        message.success(`导入成功，共处理 ${resp.data.imported} 条`);
        await reload();
      } catch (e) {
        message.error(parseApiError(e, '导入供应商失败'));
      }
      return false;
    }
  };

  useEffect(() => {
    reload();
  }, []);

  async function reload(q?: string) {
    setLoading(true);
    try {
      const resp = ensureApiOk(await listSuppliers(q ?? query));
      setSuppliers(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载供应商列表失败'));
    } finally {
      setLoading(false);
    }
  }

  function resetForm() {
    setEditingID('');
    form.setFieldsValue(EMPTY_FORM);
  }

  function onEdit(item: Supplier) {
    setEditingID(item.supplier_id);
    form.setFieldsValue({
      company_full_name: item.company_full_name,
      tax_number: item.tax_number,
      project_owner: item.project_owner,
      project_owner_phone: item.project_owner_phone,
      tech_contact: item.tech_contact,
      tech_contact_phone: item.tech_contact_phone,
      business_scope: item.business_scope
    });
  }

  async function onDelete(supplierId: string) {
    try {
      ensureApiOk(await deleteSupplier(supplierId));
      message.success('供应商已删除');
      if (editingID === supplierId) resetForm();
      await reload();
    } catch (e) {
      message.error(parseApiError(e, '删除供应商失败'));
    }
  }

  async function onSubmit(values: SupplierFormValue) {
    const payload = {
      company_full_name: values.company_full_name.trim(),
      tax_number: values.tax_number.trim().toUpperCase(),
      project_owner: values.project_owner.trim(),
      project_owner_phone: values.project_owner_phone.trim(),
      tech_contact: values.tech_contact.trim(),
      tech_contact_phone: values.tech_contact_phone.trim(),
      business_scope: values.business_scope.trim()
    };
    setSaving(true);
    try {
      if (editingID) {
        ensureApiOk(await updateSupplier(editingID, payload));
        message.success('供应商已更新');
      } else {
        ensureApiOk(await createSupplier(payload));
        message.success('供应商已创建');
      }
      resetForm();
      await reload();
    } catch (e) {
      message.error(parseApiError(e, editingID ? '更新供应商失败' : '创建供应商失败'));
    } finally {
      setSaving(false);
    }
  }

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card title="商务条款 - 供应商管理" extra={editingID ? <Tag color="blue">编辑中：{editingID}</Tag> : <Tag>新建</Tag>}>
        <Form form={form} layout="vertical" initialValues={EMPTY_FORM} onFinish={onSubmit}>
          <Space align="start" wrap style={{ width: '100%' }}>
            <Form.Item name="company_full_name" label="公司全名" rules={[{ required: true, message: '请输入公司全名' }]} style={{ width: 360 }}>
              <Input placeholder="例如：某某科技有限公司" />
            </Form.Item>
            <Form.Item
              name="tax_number"
              label="税号"
              rules={[
                { required: true, message: '请输入税号' },
                { pattern: TAX_REGEX, message: '税号格式不正确（15-20位大写字母/数字）' }
              ]}
              style={{ width: 260 }}
            >
              <Input placeholder="统一社会信用代码 / 税号" />
            </Form.Item>
            <Form.Item name="project_owner" label="项目负责人" style={{ width: 180 }}>
              <Input placeholder="负责人姓名" />
            </Form.Item>
            <Form.Item
              name="project_owner_phone"
              label="项目负责人电话"
              rules={[{ pattern: PHONE_REGEX, message: '电话格式不正确（11位手机号或区号座机）' }]}
              style={{ width: 200 }}
            >
              <Input placeholder="电话" />
            </Form.Item>
            <Form.Item name="tech_contact" label="技术接口人" style={{ width: 180 }}>
              <Input placeholder="技术联系人" />
            </Form.Item>
            <Form.Item
              name="tech_contact_phone"
              label="技术接口人电话"
              rules={[{ pattern: PHONE_REGEX, message: '电话格式不正确（11位手机号或区号座机）' }]}
              style={{ width: 200 }}
            >
              <Input placeholder="电话" />
            </Form.Item>
            <Form.Item name="business_scope" label="业务范围" rules={[{ required: true, message: '请输入业务范围' }]} style={{ minWidth: 420, flex: 1 }}>
              <Input.TextArea rows={2} placeholder="例如：云资源采购、服务器维护、IDC 服务等" />
            </Form.Item>
          </Space>
          <Space>
            <Button type="primary" htmlType="submit" loading={saving}>{editingID ? '保存修改' : '创建供应商'}</Button>
            <Button onClick={resetForm}>清空</Button>
          </Space>
        </Form>
      </Card>

      <Card
        title="供应商清单"
        extra={(
          <Space>
            <Upload {...uploadProps}>
              <Button>导入 Excel</Button>
            </Upload>
            <Button onClick={exportSupplierTemplate}>下载模板</Button>
            <Button onClick={exportSuppliers}>导出 Excel</Button>
            <Input.Search
              allowClear
              placeholder="按公司、税号、联系人、业务范围搜索"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onSearch={(v) => reload(v)}
              style={{ width: 300 }}
            />
            <Button onClick={() => reload()} loading={loading}>刷新</Button>
          </Space>
        )}
      >
        <Table<Supplier>
          rowKey="supplier_id"
          loading={loading}
          columns={columns}
          dataSource={suppliers}
          scroll={{ x: 1500 }}
          pagination={{ pageSize: 20, showSizeChanger: true }}
        />
      </Card>
    </Space>
  );
}
