import { useEffect, useState } from 'react';
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  Modal,
  Row,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  message,
  Popconfirm,
  Switch,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  archiveMetaModel,
  createMetaField,
  createMetaModel,
  createMetaReference,
  deleteMetaField,
  deleteMetaModel,
  deleteMetaReference,
  getMetaModel,
  listMetaModels,
  listMetaReferences,
  reorderMetaFields,
  updateMetaField,
  updateMetaModel,
  updateMetaReference,
  publishMetaModel,
  listMetaModelVersions,
  rollbackMetaModel
} from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { MetaField, MetaModel, MetaReference, MetaModelVersion } from '../types';

const VALUE_TYPES = ['string', 'int', 'decimal', 'bool', 'date', 'datetime', 'enum', 'json'];

export default function MetaModelPage() {
  const [models, setModels] = useState<MetaModel[]>([]);
  const [loadingModels, setLoadingModels] = useState(false);
  const [selectedModelId, setSelectedModelId] = useState<string>('');
  const [selectedModel, setSelectedModel] = useState<MetaModel | null>(null);
  const [fields, setFields] = useState<MetaField[]>([]);
  const [refs, setRefs] = useState<MetaReference[]>([]);
  const [versions, setVersions] = useState<MetaModelVersion[]>([]);
  const [loadingDetail, setLoadingDetail] = useState(false);

  const [createModelOpen, setCreateModelOpen] = useState(false);
  const [editModelOpen, setEditModelOpen] = useState(false);
  const [fieldOpen, setFieldOpen] = useState(false);
  const [refOpen, setRefOpen] = useState(false);

  const [editingField, setEditingField] = useState<MetaField | null>(null);
  const [editingRef, setEditingRef] = useState<MetaReference | null>(null);

  const [modelForm] = Form.useForm();
  const [editModelForm] = Form.useForm();
  const [fieldForm] = Form.useForm();
  const [refForm] = Form.useForm();

  useEffect(() => {
    loadModels();
  }, []);

  useEffect(() => {
    if (selectedModelId) {
      loadModelDetail(selectedModelId);
    } else {
      setSelectedModel(null);
      setFields([]);
      setRefs([]);
    }
  }, [selectedModelId]);

  async function loadModels() {
    setLoadingModels(true);
    try {
      const resp = ensureApiOk(await listMetaModels());
      const list = resp.data.list || [];
      setModels(list);
      if (!selectedModelId && list.length > 0) {
        setSelectedModelId(list[0].id);
      }
    } catch (e) {
      message.error(parseApiError(e, '加载模型列表失败'));
    } finally {
      setLoadingModels(false);
    }
  }

  async function loadModelDetail(modelId: string) {
    setLoadingDetail(true);
    try {
      const detail = ensureApiOk(await getMetaModel(modelId));
      setSelectedModel(detail.data.model);
      setFields((detail.data.fields || []).slice().sort((a, b) => a.sort_no - b.sort_no));
      const r = ensureApiOk(await listMetaReferences(modelId));
      setRefs(r.data.list || []);
      const vs = ensureApiOk(await listMetaModelVersions(modelId));
      setVersions(vs.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载模型详情失败'));
    } finally {
      setLoadingDetail(false);
    }
  }

  async function onCreateModel() {
    try {
      const values = await modelForm.validateFields();
      ensureApiOk(await createMetaModel(values));
      message.success('模型创建成功');
      setCreateModelOpen(false);
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
      setEditModelOpen(false);
      await loadModels();
      await loadModelDetail(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, '更新模型失败'));
    }
  }

  async function onArchiveModel() {
    if (!selectedModel) return;
    try {
      ensureApiOk(await archiveMetaModel(selectedModel.id));
      message.success('模型已归档');
      await loadModels();
      await loadModelDetail(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, '归档模型失败'));
    }
  }

  async function onDeleteModel() {
    if (!selectedModel) return;
    try {
      ensureApiOk(await deleteMetaModel(selectedModel.id));
      message.success('模型已删除');
      const deletedId = selectedModel.id;
      await loadModels();
      if (selectedModelId === deletedId) {
        const next = models.filter((m) => m.id !== deletedId)[0];
        setSelectedModelId(next?.id || '');
      }
    } catch (e) {
      message.error(parseApiError(e, '删除模型失败'));
    }
  }

  function openCreateField() {
    setEditingField(null);
    fieldForm.setFieldsValue({
      category: 'business',
      value_type: 'string',
      required: false,
      unique: false,
      filterable: true,
      sortable: false,
      visible: true
    });
    setFieldOpen(true);
  }

  function openEditField(row: MetaField) {
    setEditingField(row);
    fieldForm.setFieldsValue(row);
    setFieldOpen(true);
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
      setFieldOpen(false);
      fieldForm.resetFields();
      await loadModelDetail(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, editingField ? '更新属性失败' : '创建属性失败'));
    }
  }

  async function onDeleteField(row: MetaField) {
    if (!selectedModel) return;
    try {
      ensureApiOk(await deleteMetaField(selectedModel.id, row.id));
      message.success('属性删除成功');
      await loadModelDetail(selectedModel.id);
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
      message.success('排序已更新');
    } catch (e) {
      message.error(parseApiError(e, '调整排序失败'));
    }
  }

  function openCreateRef() {
    setEditingRef(null);
    refForm.setFieldsValue({ on_delete_action: 'restrict', display_fields: [] });
    setRefOpen(true);
  }

  function openEditRef(row: MetaReference) {
    setEditingRef(row);
    refForm.setFieldsValue(row);
    setRefOpen(true);
  }

  async function onSubmitRef() {
    if (!selectedModel) return;
    try {
      const values = await refForm.validateFields();
      if (editingRef) {
        ensureApiOk(await updateMetaReference(selectedModel.id, editingRef.id, values));
        message.success('关联更新成功');
      } else {
        ensureApiOk(await createMetaReference(selectedModel.id, values));
        message.success('关联创建成功');
      }
      setRefOpen(false);
      refForm.resetFields();
      await loadModelDetail(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, editingRef ? '更新关联失败' : '创建关联失败'));
    }
  }

  async function onDeleteRef(row: MetaReference) {
    if (!selectedModel) return;
    try {
      ensureApiOk(await deleteMetaReference(selectedModel.id, row.id));
      message.success('关联删除成功');
      await loadModelDetail(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, '删除关联失败'));
    }
  }



  async function onPublishModel() {
    if (!selectedModel) return;
    try {
      ensureApiOk(await publishMetaModel(selectedModel.id, { change_summary: '前端发布操作' }));
      message.success('模型发布成功');
      await loadModels();
      await loadModelDetail(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, '发布模型失败'));
    }
  }

  async function onRollbackModel(versionNo: number) {
    if (!selectedModel) return;
    try {
      ensureApiOk(await rollbackMetaModel(selectedModel.id, versionNo));
      message.success(`已回滚到版本 v${versionNo}`);
      await loadModels();
      await loadModelDetail(selectedModel.id);
    } catch (e) {
      message.error(parseApiError(e, '回滚失败'));
    }
  }

  const modelColumns: ColumnsType<MetaModel> = [
    { title: '模型编码', dataIndex: 'model_code', key: 'model_code' },
    { title: '模型名称', dataIndex: 'model_name', key: 'model_name' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (v: string) => (
        <Tag color={v === 'draft' ? 'blue' : v === 'published' ? 'green' : 'default'}>{v}</Tag>
      )
    }
  ];

  const fieldColumns: ColumnsType<MetaField> = [
    { title: '排序', dataIndex: 'sort_no', width: 70 },
    { title: '编码', dataIndex: 'field_code' },
    { title: '名称', dataIndex: 'field_name' },
    { title: '类型', dataIndex: 'value_type', width: 90 },
    {
      title: '约束',
      width: 220,
      render: (_, row) => (
        <Space size={4} wrap>
          {row.required ? <Tag color="red">必填</Tag> : null}
          {row.unique ? <Tag color="purple">唯一</Tag> : null}
          {row.filterable ? <Tag color="blue">可筛选</Tag> : null}
          {row.sortable ? <Tag color="cyan">可排序</Tag> : null}
          {row.visible ? <Tag color="green">可展示</Tag> : <Tag>隐藏</Tag>}
        </Space>
      )
    },
    {
      title: '操作',
      width: 260,
      render: (_, row) => (
        <Space>
          <Button size="small" onClick={() => onMoveField(row, -1)}>上移</Button>
          <Button size="small" onClick={() => onMoveField(row, 1)}>下移</Button>
          <Button size="small" onClick={() => openEditField(row)}>编辑</Button>
          <Popconfirm title="确认删除该属性？" onConfirm={() => onDeleteField(row)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  const refColumns: ColumnsType<MetaReference> = [
    { title: '源字段ID', dataIndex: 'source_field_id' },
    { title: '目标模型ID', dataIndex: 'target_model_id' },
    { title: '目标字段ID', dataIndex: 'target_field_id' },
    {
      title: '展示字段',
      dataIndex: 'display_fields',
      render: (v: string[]) => (v || []).join(', ')
    },
    { title: '删除策略', dataIndex: 'on_delete_action', width: 110 },
    {
      title: '操作',
      width: 160,
      render: (_, row) => (
        <Space>
          <Button size="small" onClick={() => openEditRef(row)}>编辑</Button>
          <Popconfirm title="确认删除该关联？" onConfirm={() => onDeleteRef(row)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      )
    }
  ];


  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Row gutter={16}>
        <Col span={8}>
          <Card
            title="模型列表"
            extra={<Button type="primary" onClick={() => setCreateModelOpen(true)}>新建模型</Button>}
            loading={loadingModels}
          >
            <Table
              rowKey="id"
              size="small"
              pagination={false}
              columns={modelColumns}
              dataSource={models}
              rowSelection={{
                type: 'radio',
                selectedRowKeys: selectedModelId ? [selectedModelId] : [],
                onChange: (keys) => setSelectedModelId(String(keys[0] || ''))
              }}
            />
          </Card>
        </Col>
        <Col span={16}>
          <Card
            title={selectedModel ? `模型详情：${selectedModel.model_name}` : '模型详情'}
            loading={loadingDetail}
            extra={selectedModel ? (
              <Space>
                <Button onClick={() => {
                  editModelForm.setFieldsValue({ model_name: selectedModel.model_name, description: selectedModel.description });
                  setEditModelOpen(true);
                }}>编辑模型</Button>
                <Button type="primary" onClick={onPublishModel}>发布</Button>
                <Button onClick={onArchiveModel}>归档</Button>
                <Popconfirm title="仅草稿且无记录模型可删除，确认删除？" onConfirm={onDeleteModel}>
                  <Button danger>删除模型</Button>
                </Popconfirm>
              </Space>
            ) : null}
          >
            {!selectedModel ? <Typography.Text type="secondary">请选择左侧模型</Typography.Text> : (
              <Space direction="vertical" style={{ width: '100%' }}>
                <Typography.Text>编码：{selectedModel.model_code}</Typography.Text>
                <Typography.Text>状态：{selectedModel.status}</Typography.Text>
                <Typography.Text>描述：{selectedModel.description || '-'}</Typography.Text>
              </Space>
            )}
          </Card>



          <Card
            style={{ marginTop: 16 }}
            title="版本管理"
          >
            <Table
              rowKey="id"
              size="small"
              pagination={false}
              dataSource={versions}
              columns={[
                { title: '版本号', dataIndex: 'version_no', render: (v:number) => `v${v}` },
                { title: '发布时间', dataIndex: 'published_at' },
                { title: '变更说明', dataIndex: 'change_summary', render: (v:string) => v || '-' },
                {
                  title: '操作',
                  width: 120,
                  render: (_:unknown, row: MetaModelVersion) => (
                    <Popconfirm title={`确认回滚到 v${row.version_no} ?`} onConfirm={() => onRollbackModel(row.version_no)}>
                      <Button size="small">回滚</Button>
                    </Popconfirm>
                  )
                }
              ]}
            />
          </Card>

          <Card
            style={{ marginTop: 16 }}
            title="属性管理"
            extra={selectedModel ? <Button type="primary" onClick={openCreateField}>新增属性</Button> : null}
          >
            <Table rowKey="id" size="small" columns={fieldColumns} dataSource={fields} pagination={false} />
          </Card>

          <Card
            style={{ marginTop: 16 }}
            title="关联管理"
            extra={selectedModel ? <Button type="primary" onClick={openCreateRef}>新增关联</Button> : null}
          >
            <Table rowKey="id" size="small" columns={refColumns} dataSource={refs} pagination={false} />
          </Card>
        </Col>
      </Row>

      <Modal title="新建模型" open={createModelOpen} onCancel={() => setCreateModelOpen(false)} onOk={onCreateModel}>
        <Form layout="vertical" form={modelForm}>
          <Form.Item name="model_code" label="模型编码" rules={[{ required: true }]}>
            <Input placeholder="如 host_meta" />
          </Form.Item>
          <Form.Item name="model_name" label="模型名称" rules={[{ required: true }]}>
            <Input placeholder="如 主机元数据" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="编辑模型" open={editModelOpen} onCancel={() => setEditModelOpen(false)} onOk={onUpdateModel}>
        <Form layout="vertical" form={editModelForm}>
          <Form.Item name="model_name" label="模型名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title={editingField ? '编辑属性' : '新增属性'} open={fieldOpen} onCancel={() => setFieldOpen(false)} onOk={onSubmitField} width={680}>
        <Form layout="vertical" form={fieldForm}>
          <Row gutter={12}>
            <Col span={12}>
              <Form.Item name="field_code" label="属性编码" rules={[{ required: true }]}>
                <Input disabled={!!editingField} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="field_name" label="属性名称" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={12}>
            <Col span={8}>
              <Form.Item name="category" label="分类"><Input /></Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="value_type" label="值类型" rules={[{ required: true }]}>
                <Select options={VALUE_TYPES.map((v) => ({ label: v, value: v }))} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="default_value" label="默认值"><Input /></Form.Item>
            </Col>
          </Row>
          <Form.Item name="validation_rule" label="校验规则"><Input /></Form.Item>
          <Space wrap>
            <Form.Item name="required" valuePropName="checked" label="必填"><Switch /></Form.Item>
            <Form.Item name="unique" valuePropName="checked" label="唯一"><Switch /></Form.Item>
            <Form.Item name="filterable" valuePropName="checked" label="可筛选"><Switch /></Form.Item>
            <Form.Item name="sortable" valuePropName="checked" label="可排序"><Switch /></Form.Item>
            <Form.Item name="visible" valuePropName="checked" label="可展示"><Switch /></Form.Item>
          </Space>
        </Form>
      </Modal>

      <Modal title={editingRef ? '编辑关联' : '新增关联'} open={refOpen} onCancel={() => setRefOpen(false)} onOk={onSubmitRef}>
        <Form layout="vertical" form={refForm}>
          <Form.Item name="source_field_id" label="源字段ID" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="target_model_id" label="目标模型ID" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="target_field_id" label="目标字段ID" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="display_fields" label="展示字段ID（多选）">
            <Select mode="tags" tokenSeparators={[',']} />
          </Form.Item>
          <Form.Item name="on_delete_action" label="删除策略">
            <Select options={[{ label: 'restrict', value: 'restrict' }, { label: 'set_null', value: 'set_null' }, { label: 'cascade', value: 'cascade' }]} />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
