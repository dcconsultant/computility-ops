import { useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import {
  Alert,
  Button,
  Card,
  Col,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Tree,
  Typography,
  Upload,
  message
} from 'antd';
import { MenuUnfoldOutlined, MenuFoldOutlined, PlusOutlined, SettingOutlined, UploadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import {
  cloneMetaModel,
  createMetaField,
  createMetaModel,
  createMetaRecord,
  deleteMetaField,
  deleteMetaModel,
  deleteMetaRecord,
  getMetaModel,
  listMetaModelVersions,
  listMetaModels,
  listMetaRecords,
  exportMetaRecordTemplate,
  importMetaRecords,
  publishMetaModel,
  reorderMetaFields,
  updateMetaField,
  updateMetaModel,
  updateMetaRecord
} from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { MetaField, MetaModel, MetaModelVersion, MetaRecord } from '../types';

const VALUE_TYPES = ['string', 'int', 'decimal', 'bool', 'date', 'datetime', 'enum', 'json'];

type ViewMode = 'data' | 'config';

export default function MetaModelPage() {
  const [models, setModels] = useState<MetaModel[]>([]);
  const [selectedModelId, setSelectedModelId] = useState<string>('');
  const [selectedModel, setSelectedModel] = useState<MetaModel | null>(null);
  const [fields, setFields] = useState<MetaField[]>([]);
  const [versions, setVersions] = useState<MetaModelVersion[]>([]);
  const [records, setRecords] = useState<MetaRecord[]>([]);
  const [mode, setMode] = useState<ViewMode>('data');
  const [loading, setLoading] = useState(false);
  const [navCollapsed, setNavCollapsed] = useState(false);
  const [importing, setImporting] = useState(false);
  const [importResultOpen, setImportResultOpen] = useState(false);
  const [importSummary, setImportSummary] = useState<{ total: number; success: number; failed: number }>({ total: 0, success: 0, failed: 0 });
  const [importErrors, setImportErrors] = useState<Array<{ row: number; error: string }>>([]);

  const [modelModalOpen, setModelModalOpen] = useState(false);
  const [editModelModalOpen, setEditModelModalOpen] = useState(false);
  const [fieldModalOpen, setFieldModalOpen] = useState(false);
  const [recordModalOpen, setRecordModalOpen] = useState(false);
  const [cloneModalOpen, setCloneModalOpen] = useState(false);

  const [editingField, setEditingField] = useState<MetaField | null>(null);
  const [editingRecord, setEditingRecord] = useState<MetaRecord | null>(null);

  const [modelForm] = Form.useForm();
  const [editModelForm] = Form.useForm();
  const [fieldForm] = Form.useForm();
  const [recordForm] = Form.useForm();
  const [cloneForm] = Form.useForm();
  const [filterForm] = Form.useForm();
  const fieldValueType = Form.useWatch('value_type', fieldForm);
  const [searchParams, setSearchParams] = useSearchParams();
  const [filteredRecords, setFilteredRecords] = useState<MetaRecord[]>([]);

  useEffect(() => {
    const qMode = searchParams.get('mode');
    if (qMode === 'config' || qMode === 'data') {
      setMode(qMode);
    }
    const qModel = searchParams.get('modelId');
    if (qModel) {
      setSelectedModelId(qModel);
    }
    loadModels();
  }, []);

  useEffect(() => {
    if (selectedModelId) {
      setSearchParams({ modelId: selectedModelId, mode });
      if (mode === 'data') {
        loadRecordView(selectedModelId);
      } else {
        loadConfigView(selectedModelId);
      }
    }
  }, [selectedModelId, mode]);

  useEffect(() => {
    setFilteredRecords(records);
  }, [records]);

  async function loadModels() {
    try {
      const resp = ensureApiOk(await listMetaModels());
      const list = resp.data.list || [];
      setModels(list);
      const cached = localStorage.getItem('meta:selectedModelId');
      const queryModel = searchParams.get('modelId');
      const chosen = list.find((m) => m.id === queryModel)?.id || list.find((m) => m.id === cached)?.id || list[0]?.id || '';
      if (!selectedModelId && chosen) {
        setSelectedModelId(chosen);
      }
    } catch (e) {
      message.error(parseApiError(e, '加载模型列表失败'));
    }
  }

  async function loadRecordView(modelId: string) {
    setLoading(true);
    try {
      const resp = ensureApiOk(await listMetaRecords(modelId));
      setSelectedModel(resp.data.model);
      setFields((resp.data.fields || []).slice().sort((a, b) => a.sort_no - b.sort_no));
      setRecords(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载模型数据失败'));
    } finally {
      setLoading(false);
    }
  }

  async function loadConfigView(modelId: string) {
    setLoading(true);
    try {
      const detail = ensureApiOk(await getMetaModel(modelId));
      setSelectedModel(detail.data.model);
      setFields((detail.data.fields || []).slice().sort((a, b) => a.sort_no - b.sort_no));
      const vs = ensureApiOk(await listMetaModelVersions(modelId));
      setVersions(vs.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载模型配置失败'));
    } finally {
      setLoading(false);
    }
  }

  const dataColumns: ColumnsType<MetaRecord> = useMemo(() => {
    const visibleFields = fields.filter((f) => f.visible);
    const fieldCols = visibleFields.map((f) => ({
      title: f.field_name,
      key: f.id,
      width: Math.max((f.field_name?.length || 6) * 18, 140),
      dataIndex: ['data', f.field_code],
      sorter: f.sortable
        ? (a: MetaRecord, b: MetaRecord) => String(a.data?.[f.field_code] ?? '').localeCompare(String(b.data?.[f.field_code] ?? ''), 'zh')
        : undefined,
      render: (v: any) => (typeof v === 'object' ? JSON.stringify(v) : String(v ?? '-'))
    }));

    return [
      ...fieldCols,
      {
        title: '更新时间',
        dataIndex: 'updated_at',
        width: 180
      },
      {
        title: '操作',
        width: 180,
        render: (_, row) => (
          <Space>
            <Button size="small" onClick={() => openEditRecord(row)}>编辑</Button>
            <Popconfirm title="确认删除该记录?" onConfirm={() => onDeleteRecord(row.id)}>
              <Button size="small" danger>删除</Button>
            </Popconfirm>
          </Space>
        )
      }
    ];
  }, [fields]);

  const fieldColumns: ColumnsType<MetaField> = [
    { title: '排序', dataIndex: 'sort_no', width: 70 },
    { title: '编码', dataIndex: 'field_code' },
    { title: '名称', dataIndex: 'field_name' },
    { title: '类型', dataIndex: 'value_type', width: 90 },
    {
      title: '约束',
      render: (_, row) => (
        <Space size={4} wrap>
          {row.required ? <Tag color="red">必填</Tag> : null}
          {row.unique ? <Tag color="purple">唯一</Tag> : null}
          {row.visible ? <Tag color="green">展示</Tag> : <Tag>隐藏</Tag>}
        </Space>
      )
    },
    {
      title: '操作',
      width: 240,
      render: (_, row) => (
        <Space>
          <Button size="small" onClick={() => onMoveField(row, -1)}>上移</Button>
          <Button size="small" onClick={() => onMoveField(row, 1)}>下移</Button>
          <Button size="small" onClick={() => openEditField(row)}>编辑</Button>
          <Popconfirm title="确认删除该属性?" onConfirm={() => onDeleteField(row.id)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  function openCreateRecord() {
    if (!selectedModel) return;
    if (selectedModel.status !== 'published') {
      message.warning('请先发布模型版本后再录入数据');
      return;
    }
    setEditingRecord(null);
    recordForm.resetFields();
    const init: Record<string, any> = {};
    fields.forEach((f) => {
      if (f.default_value) init[f.field_code] = f.default_value;
    });
    recordForm.setFieldsValue(init);
    setRecordModalOpen(true);
  }

  function openEditRecord(row: MetaRecord) {
    setEditingRecord(row);
    recordForm.setFieldsValue(row.data || {});
    setRecordModalOpen(true);
  }

  async function onSubmitRecord() {
    if (!selectedModel) return;
    try {
      const values = await recordForm.validateFields();
      if (editingRecord) {
        ensureApiOk(await updateMetaRecord(selectedModel.id, editingRecord.id, { data: values }));
        message.success('记录更新成功');
      } else {
        ensureApiOk(await createMetaRecord(selectedModel.id, { data: values }));
        message.success('记录创建成功');
      }
      setRecordModalOpen(false);
      await loadRecordView(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, editingRecord ? '更新记录失败' : '创建记录失败'));
    }
  }

  async function onDeleteRecord(recordId: string) {
    if (!selectedModel) return;
    try {
      ensureApiOk(await deleteMetaRecord(selectedModel.id, recordId));
      message.success('记录已删除');
      await loadRecordView(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, '删除记录失败'));
    }
  }

  async function onCreateModel() {
    try {
      const values = await modelForm.validateFields();
      ensureApiOk(await createMetaModel(values));
      message.success('模型创建成功');
      setModelModalOpen(false);
      modelForm.resetFields();
      await loadModels();
    } catch (e) {
      message.error(parseApiError(e, '创建模型失败'));
    }
  }

  async function onUpdateModel() {
    if (!selectedModel) return;
    try {
      const values = await editModelForm.validateFields();
      ensureApiOk(await updateMetaModel(selectedModel.id, values));
      message.success('模型更新成功');
      setEditModelModalOpen(false);
      await loadModels();
      await loadConfigView(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, '更新模型失败'));
    }
  }

  async function onDeleteModel() {
    if (!selectedModel) return;
    try {
      ensureApiOk(await deleteMetaModel(selectedModel.id));
      message.success('模型已删除');
      setSelectedModelId('');
      await loadModels();
    } catch (e) {
      message.error(parseApiError(e, '删除模型失败'));
    }
  }

  function openCloneModel() {
    if (!selectedModel) return;
    cloneForm.setFieldsValue({
      model_code: `${selectedModel.model_code}_copy`,
      model_name: `${selectedModel.model_name}-副本`,
      description: selectedModel.description || ''
    });
    setCloneModalOpen(true);
  }

  async function onCloneModel() {
    if (!selectedModel) return;
    try {
      const values = await cloneForm.validateFields();
      const resp = ensureApiOk(await cloneMetaModel(selectedModel.id, values));
      message.success('模型克隆成功');
      setCloneModalOpen(false);
      await loadModels();
      setSelectedModelId(resp.data.id);
      setMode('config');
      localStorage.setItem('meta:selectedModelId', resp.data.id);
    } catch (e) {
      message.error(parseApiError(e, '模型克隆失败'));
    }
  }

  function openCreateField() {
    setEditingField(null);
    fieldForm.setFieldsValue({
      value_type: 'string',
      category: 'business',
      required: false,
      unique: false,
      filterable: true,
      sortable: false,
      visible: true
    });
    setFieldModalOpen(true);
  }

  function openEditField(row: MetaField) {
    setEditingField(row);
    fieldForm.setFieldsValue(row);
    setFieldModalOpen(true);
  }

  async function onSubmitField() {
    if (!selectedModel) return;
    try {
      const values = await fieldForm.validateFields();
      if (editingField) {
        ensureApiOk(await updateMetaField(selectedModel.id, editingField.id, values));
        message.success('属性更新成功');
      } else {
        ensureApiOk(await createMetaField(selectedModel.id, values));
        message.success('属性创建成功');
      }
      setFieldModalOpen(false);
      await loadConfigView(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, editingField ? '更新属性失败' : '创建属性失败'));
    }
  }

  async function onDeleteField(fieldId: string) {
    if (!selectedModel) return;
    try {
      ensureApiOk(await deleteMetaField(selectedModel.id, fieldId));
      message.success('属性删除成功');
      await loadConfigView(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, '删除属性失败'));
    }
  }

  async function onMoveField(row: MetaField, direction: -1 | 1) {
    if (!selectedModel) return;
    const idx = fields.findIndex((f) => f.id === row.id);
    const target = idx + direction;
    if (idx < 0 || target < 0 || target >= fields.length) return;
    const next = [...fields];
    [next[idx], next[target]] = [next[target], next[idx]];
    try {
      ensureApiOk(await reorderMetaFields(selectedModel.id, next.map((f) => f.id)));
      setFields(next.map((f, i) => ({ ...f, sort_no: i + 1 })));
    } catch (e) {
      message.error(parseApiError(e, '调整排序失败'));
    }
  }

  async function onPublishModel() {
    if (!selectedModel) return;
    try {
      ensureApiOk(await publishMetaModel(selectedModel.id, { change_summary: '前端发布操作' }));
      message.success('模型发布成功');
      await loadModels();
      await loadConfigView(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, '发布模型失败'));
    }
  }

  function onExportTemplate() {
    if (!selectedModel) return;
    exportMetaRecordTemplate(selectedModel.id);
  }

  async function onImportRecords(file: File) {
    if (!selectedModel) return false;
    try {
      setImporting(true);
      const resp = ensureApiOk(await importMetaRecords(selectedModel.id, file));
      setImportSummary({ total: resp.data.total, success: resp.data.success, failed: resp.data.failed });
      setImportErrors((resp.data.errors || []).map((it: any) => ({ row: Number(it.row || 0), error: String(it.error || '') })));
      setImportResultOpen(true);
      message.success(`导入完成：成功 ${resp.data.success}，失败 ${resp.data.failed}`);
      await loadRecordView(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, '导入失败'));
    } finally {
      setImporting(false);
    }
    return false;
  }

  function applyFilters() {
    const values = filterForm.getFieldsValue();
    const filtered = records.filter((rec) => {
      return fields.every((f) => {
        if (!f.filterable) return true;
        const cond = values[f.field_code];
        if (cond === undefined || cond === null || cond === '') return true;
        const val = rec.data?.[f.field_code];
        if (val === undefined || val === null) return false;
        if (f.value_type === 'string' || f.value_type === 'date' || f.value_type === 'datetime') {
          return String(val).toLowerCase().includes(String(cond).toLowerCase());
        }
        return String(val) === String(cond);
      });
    });
    setFilteredRecords(filtered);
  }

  function resetFilters() {
    filterForm.resetFields();
    setFilteredRecords(records);
  }

  return (
    <Row gutter={12} style={{ width: '100%', margin: 0 }}>
      <Col flex={navCollapsed ? '44px' : '16ch'} style={{ paddingLeft: 0 }}>
        <Card
          title="模型导航"
          extra={
            <Space>
              {!navCollapsed ? <Button type={mode === 'config' ? 'primary' : 'default'} icon={<SettingOutlined />} onClick={() => setMode('config')} /> : null}
              {!navCollapsed ? <Button icon={<PlusOutlined />} onClick={() => setModelModalOpen(true)} /> : null}
              <Button icon={navCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />} onClick={() => setNavCollapsed((v) => !v)} />
            </Space>
          }
        >
          {!navCollapsed ? <Tree
            selectedKeys={selectedModelId ? [selectedModelId] : []}
            treeData={models.map((m) => ({
              key: m.id,
              title: navCollapsed ? m.model_name.slice(0, 1) : (
                <Space>
                  <span>{m.model_name}</span>
                  <Tag color={m.status === 'published' ? 'green' : m.status === 'draft' ? 'blue' : 'default'}>{m.status}</Tag>
                </Space>
              )
            }))}
            onSelect={(keys) => {
              const id = String(keys[0] || '');
              if (!id) return;
              setSelectedModelId(id);
              setMode('data');
              localStorage.setItem('meta:selectedModelId', id);
            }}
          /> : null}
        </Card>
      </Col>

      <Col flex="auto" style={{ paddingRight: 0 }}>
        {mode === 'data' ? (
          <Card
            loading={loading}
            title={selectedModel ? `数据管理:${selectedModel.model_name} (${selectedModel.model_code})` : '数据管理'}
            extra={
              <Space>
                <Button icon={<SettingOutlined />} onClick={() => setMode('config')}>模型配置</Button>
                <Button onClick={onExportTemplate}>导入模板</Button>
                <Upload beforeUpload={(file) => onImportRecords(file as File)} showUploadList={false} disabled={!selectedModel || selectedModel.status !== 'published'}>
                  <Button icon={<UploadOutlined />} loading={importing} disabled={!selectedModel || selectedModel.status !== 'published'}>Excel导入</Button>
                </Upload>
                <Button type="primary" onClick={openCreateRecord} disabled={!selectedModel || selectedModel.status !== 'published'}>新增记录</Button>
              </Space>
            }
          >
            {selectedModel?.status !== 'published' ? <Alert type="warning" showIcon message="当前模型未发布,暂不可写入数据。" style={{ marginBottom: 12 }} /> : null}
            <Form form={filterForm} layout="vertical" style={{ marginBottom: 12 }}>
              <Row gutter={12}>
                {fields.filter((f) => f.filterable).map((f) => (
                  <Col span={8} key={f.id}>
                    <Form.Item name={f.field_code} label={f.field_name}>
                      {f.value_type === 'enum' ? (
                        <Select allowClear options={(f.enum_options || []).filter((v) => !v.disabled).map((v) => ({ label: v.label, value: v.value }))} />
                      ) : f.value_type === 'bool' ? (
                        <Select allowClear options={[{ label: 'true', value: true }, { label: 'false', value: false }]} />
                      ) : (
                        <Input allowClear placeholder={`按${f.field_name}筛选`} />
                      )}
                    </Form.Item>
                  </Col>
                ))}
              </Row>
              <Space style={{ marginBottom: 8 }}>
                <Button type="primary" onClick={applyFilters}>查询</Button>
                <Button onClick={resetFilters}>重置</Button>
              </Space>
            </Form>
            <Table rowKey="id" columns={dataColumns} dataSource={filteredRecords} scroll={{ x: 'max-content' }} />
          </Card>
        ) : (
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Card
              loading={loading}
              title={selectedModel ? `模型配置:${selectedModel.model_name}` : '模型配置'}
              extra={selectedModel ? (
                <Space>
                  <Button onClick={() => {
                    editModelForm.setFieldsValue({ model_name: selectedModel.model_name, description: selectedModel.description });
                    setEditModelModalOpen(true);
                  }}>保存</Button>
                  <Button type="primary" onClick={onPublishModel}>发布</Button>
                  <Popconfirm title="确认删除模型?" onConfirm={onDeleteModel}>
                    <Button danger>删除模型</Button>
                  </Popconfirm>
                  <Button onClick={openCloneModel}>克隆模型</Button>
                </Space>
              ) : null}
            >
              {selectedModel ? (
                <Descriptions bordered size="small" column={2}>
                  <Descriptions.Item label="模型编码">{selectedModel.model_code}</Descriptions.Item>
                  <Descriptions.Item label="状态">{selectedModel.status}</Descriptions.Item>
                  <Descriptions.Item label="当前版本">v{selectedModel.current_version}</Descriptions.Item>
                  <Descriptions.Item label="描述">{selectedModel.description || '-'}</Descriptions.Item>
                </Descriptions>
              ) : <Typography.Text type="secondary">请选择模型</Typography.Text>}
            </Card>

            <Card title="属性管理" extra={selectedModel ? <Button onClick={openCreateField}>新增属性</Button> : null}>
              <Table rowKey="id" columns={fieldColumns} dataSource={fields} />
            </Card>

            <Card title="发布版本">
              <Table
                rowKey="id"
                dataSource={versions}
                columns={[
                  { title: '版本', dataIndex: 'version_no', render: (v: number) => `v${v}` },
                  { title: '发布时间', dataIndex: 'published_at' },
                  { title: '说明', dataIndex: 'change_summary', render: (v: string) => v || '-' }
                ]}
              />
            </Card>
          </Space>
        )}
      </Col>

      <Modal title="新建模型" open={modelModalOpen} onCancel={() => setModelModalOpen(false)} onOk={onCreateModel}>
        <Form layout="vertical" form={modelForm}>
          <Form.Item name="model_code" label="模型编码" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="model_name" label="模型名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>

      <Modal title="编辑模型" open={editModelModalOpen} onCancel={() => setEditModelModalOpen(false)} onOk={onUpdateModel}>
        <Form layout="vertical" form={editModelForm}>
          <Form.Item name="model_name" label="模型名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>

      <Modal title={editingField ? '编辑属性' : '新增属性'} open={fieldModalOpen} onCancel={() => setFieldModalOpen(false)} onOk={onSubmitField}>
        <Form layout="vertical" form={fieldForm}>
          <Form.Item name="field_code" label="属性编码" rules={[{ required: true }]}><Input disabled={!!editingField} /></Form.Item>
          <Form.Item name="field_name" label="属性名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="category" label="分类"><Input /></Form.Item>
          <Form.Item name="value_type" label="值类型" rules={[{ required: true }]}>
            <Select options={VALUE_TYPES.map((v) => ({ label: v, value: v }))} />
          </Form.Item>

          <Space wrap>
            <Form.Item name="required" valuePropName="checked" label="是否必填"><Switch /></Form.Item>
            <Form.Item name="unique" valuePropName="checked" label="是否唯一"><Switch /></Form.Item>
            <Form.Item name="visible" valuePropName="checked" label="默认展示"><Switch /></Form.Item>
            <Form.Item name="filterable" valuePropName="checked" label="支持搜索"><Switch /></Form.Item>
            <Form.Item name="sortable" valuePropName="checked" label="支持排序"><Switch /></Form.Item>
          </Space>

          {fieldValueType === 'enum' ? (
            <Form.List name="enum_options" rules={[{ validator: async (_, v) => { if (!v || v.length === 0) throw new Error('请至少配置1个枚举项'); } }]}>
              {(items, { add, remove }, { errors }) => (
                <Space direction="vertical" style={{ width: '100%' }}>
                  {items.map((item) => (
                    <Row key={item.key} gutter={8}>
                      <Col span={10}><Form.Item name={[item.name, 'value']} rules={[{ required: true, message: 'value必填' }]}><Input placeholder="value" /></Form.Item></Col>
                      <Col span={10}><Form.Item name={[item.name, 'label']} rules={[{ required: true, message: 'label必填' }]}><Input placeholder="label" /></Form.Item></Col>
                      <Col span={4}><Button danger onClick={() => remove(item.name)}>删除</Button></Col>
                    </Row>
                  ))}
                  <Button onClick={() => add({ value: '', label: '', disabled: false })}>新增枚举项</Button>
                  <Form.ErrorList errors={errors} />
                </Space>
              )}
            </Form.List>
          ) : null}
        </Form>
      </Modal>

      <Modal title={editingRecord ? '编辑记录' : '新增记录'} open={recordModalOpen} onCancel={() => setRecordModalOpen(false)} onOk={onSubmitRecord} width={760}>
        <Form layout="vertical" form={recordForm}>
          {fields.map((f) => {
            if (f.value_type === 'bool') {
              return (
                <Form.Item key={f.id} name={f.field_code} label={f.field_name} rules={f.required ? [{ required: true }] : []}>
                  <Select options={[{ label: 'true', value: true }, { label: 'false', value: false }]} />
                </Form.Item>
              );
            }
            if (f.value_type === 'enum') {
              return (
                <Form.Item key={f.id} name={f.field_code} label={f.field_name} rules={f.required ? [{ required: true }] : []}>
                  <Select options={(f.enum_options || []).filter((v) => !v.disabled).map((v) => ({ label: v.label, value: v.value }))} />
                </Form.Item>
              );
            }
            if (f.value_type === 'int' || f.value_type === 'decimal') {
              return (
                <Form.Item key={f.id} name={f.field_code} label={f.field_name} rules={f.required ? [{ required: true }] : []}>
                  <InputNumber style={{ width: '100%' }} />
                </Form.Item>
              );
            }
            return (
              <Form.Item key={f.id} name={f.field_code} label={f.field_name} rules={f.required ? [{ required: true }] : []}>
                <Input />
              </Form.Item>
            );
          })}
        </Form>
      </Modal>

      <Modal title="克隆模型" open={cloneModalOpen} onCancel={() => setCloneModalOpen(false)} onOk={onCloneModel}>
        <Form layout="vertical" form={cloneForm}>
          <Form.Item name="model_code" label="模型编码" rules={[{ required: true, message: '请输入模型编码' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="model_name" label="模型名称" rules={[{ required: true, message: '请输入模型名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="Excel 导入结果"
        open={importResultOpen}
        onCancel={() => setImportResultOpen(false)}
        footer={<Button type="primary" onClick={() => setImportResultOpen(false)}>知道了</Button>}
        width={820}
      >
        <Space direction="vertical" style={{ width: '100%' }} size={12}>
          <Alert
            type={importSummary.failed > 0 ? 'warning' : 'success'}
            showIcon
            message={`总计 ${importSummary.total} 条，成功 ${importSummary.success} 条，失败 ${importSummary.failed} 条`}
          />
          {importSummary.failed > 0 ? (
            <Table
              size="small"
              rowKey={(r) => `${r.row}-${r.error}`}
              dataSource={importErrors}
              pagination={{ pageSize: 10 }}
              columns={[
                { title: '失败行', dataIndex: 'row', width: 100 },
                { title: '错误原因', dataIndex: 'error' }
              ]}
            />
          ) : null}
        </Space>
      </Modal>
    </Row>
  );
}
