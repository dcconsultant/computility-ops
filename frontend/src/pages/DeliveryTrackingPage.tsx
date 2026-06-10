import { useEffect, useState } from 'react';
import { Button, Card, DatePicker, Form, Input, InputNumber, Popconfirm, Select, Space, Table, Tabs, Tag, Upload, message } from 'antd';
import { DownloadOutlined, UploadOutlined } from '@ant-design/icons';
import type { Dayjs } from 'dayjs';
import type { UploadProps } from 'antd';
import dayjs from 'dayjs';
import {
  createAccessoryArrival,
  createArrivalPlan,
  createDeviceArrival,
  deleteAccessoryArrival,
  deleteArrivalPlan,
  deleteDeviceArrival,
  exportAccessoryArrivalTemplate,
  exportAccessoryArrivals,
  exportArrivalPlanTemplate,
  exportArrivalPlans,
  exportDeviceArrivalTemplate,
  exportDeviceArrivals,
  importAccessoryArrivals,
  importArrivalPlans,
  importDeviceArrivals,
  listAccessoryArrivals,
  listArrivalPlans,
  listDeviceArrivals,
  updateAccessoryArrival,
  updateArrivalPlan,
  updateDeviceArrival
} from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { AccessoryArrivalRecord, ArrivalPlan, DeviceArrivalRecord } from '../types';

const planCategories = ['服务器', '网络设备', '耗材及配件'];
const deviceCategories = ['服务器', '网络设备'];
const dateTimeFormat = 'YYYY-MM-DD HH:mm:ss';

type PlanFormValue = Omit<ArrivalPlan, 'plan_id' | 'created_at' | 'updated_at' | 'estimated_arrival_time'> & {
  estimated_arrival_time: Dayjs;
};

type DeviceFormValue = Omit<DeviceArrivalRecord, 'record_id' | 'created_at' | 'updated_at' | 'srm_requirement_submitted_at' | 'actual_arrival_time'> & {
  srm_requirement_submitted_at: Dayjs;
  actual_arrival_time: Dayjs;
};

type AccessoryFormValue = Omit<AccessoryArrivalRecord, 'record_id' | 'created_at' | 'updated_at' | 'srm_requirement_submitted_at' | 'arrival_time'> & {
  srm_requirement_submitted_at: Dayjs;
  arrival_time: Dayjs;
};

export default function DeliveryTrackingPage() {
  return (
    <Tabs
      defaultActiveKey="plans"
      items={[
        { key: 'plans', label: '到货计划', children: <ArrivalPlanTab /> },
        { key: 'devices', label: '服务器&网络设备到货记录', children: <DeviceArrivalTab /> },
        { key: 'accessories', label: '配件&耗材到货记录', children: <AccessoryArrivalTab /> }
      ]}
    />
  );
}

function ArrivalPlanTab() {
  const [form] = Form.useForm<PlanFormValue>();
  const [filterForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editingID, setEditingID] = useState('');
  const [rows, setRows] = useState<ArrivalPlan[]>([]);

  async function reload() {
    setLoading(true);
    try {
      const filters = buildQuery(filterForm.getFieldsValue());
      const resp = ensureApiOk(await listArrivalPlans(filters));
      setRows(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载到货计划失败'));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    reload();
  }, []);

  function resetForm() {
    setEditingID('');
    form.resetFields();
  }

  async function onSubmit(values: PlanFormValue) {
    setSaving(true);
    try {
      const payload = {
        category: values.category,
        material_code: trim(values.material_code),
        material_name: trim(values.material_name),
        quantity: Number(values.quantity || 0),
        receiving_address: trim(values.receiving_address),
        supplier: trim(values.supplier),
        order_no: trim(values.order_no),
        asset_code_range: trim(values.asset_code_range),
        estimated_arrival_time: formatDateTime(values.estimated_arrival_time),
        remark: trim(values.remark)
      };
      if (editingID) {
        ensureApiOk(await updateArrivalPlan(editingID, payload));
        message.success('到货计划已更新');
      } else {
        ensureApiOk(await createArrivalPlan(payload));
        message.success('到货计划已创建');
      }
      resetForm();
      await reload();
    } catch (e) {
      message.error(parseApiError(e, editingID ? '更新到货计划失败' : '创建到货计划失败'));
    } finally {
      setSaving(false);
    }
  }

  function onEdit(row: ArrivalPlan) {
    setEditingID(row.plan_id);
    form.setFieldsValue({
      ...row,
      estimated_arrival_time: parseDateTime(row.estimated_arrival_time)
    });
  }

  async function onDelete(planId: string) {
    try {
      ensureApiOk(await deleteArrivalPlan(planId));
      message.success('到货计划已删除');
      if (editingID === planId) resetForm();
      await reload();
    } catch (e) {
      message.error(parseApiError(e, '删除到货计划失败'));
    }
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card title="到货计划 - 新建/编辑" extra={editingID ? <Tag color="blue">编辑中：{editingID}</Tag> : <Tag>新建</Tag>}>
        <Form form={form} layout="vertical" onFinish={onSubmit}>
          <Space wrap>
            <Form.Item name="category" label="类别" rules={[{ required: true, message: '请选择类别' }]} style={{ width: 180 }}>
              <Select options={planCategories.map((value) => ({ value, label: value }))} />
            </Form.Item>
            <Form.Item name="material_code" label="物料代码" rules={[{ required: true, message: '请输入物料代码' }]} style={{ width: 220 }}>
              <Input />
            </Form.Item>
            <Form.Item name="material_name" label="物料名称" rules={[{ required: true, message: '请输入物料名称' }]} style={{ width: 240 }}>
              <Input />
            </Form.Item>
            <Form.Item name="quantity" label="数量（台）" rules={[{ required: true, message: '请输入数量' }]} style={{ width: 160 }}>
              <InputNumber min={1} precision={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="supplier" label="供应商" rules={[{ required: true, message: '请输入供应商' }]} style={{ width: 220 }}>
              <Input />
            </Form.Item>
            <Form.Item name="order_no" label="订单号" rules={[{ required: true, message: '请输入订单号' }]} style={{ width: 220 }}>
              <Input />
            </Form.Item>
            <Form.Item name="estimated_arrival_time" label="预估到货时间" rules={[{ required: true, message: '请选择预估到货时间' }]}>
              <DatePicker showTime format={dateTimeFormat} />
            </Form.Item>
          </Space>
          <Space wrap>
            <Form.Item name="receiving_address" label="收货地址" rules={[{ required: true, message: '请输入收货地址' }]} style={{ width: 360 }}>
              <Input />
            </Form.Item>
            <Form.Item name="asset_code_range" label="资产编码区间" style={{ width: 260 }}>
              <Input />
            </Form.Item>
            <Form.Item name="remark" label="备注" style={{ width: 360 }}>
              <Input />
            </Form.Item>
          </Space>
          <Space>
            <Button type="primary" htmlType="submit" loading={saving}>{editingID ? '保存更新' : '创建计划'}</Button>
            <Button onClick={resetForm} disabled={saving}>重置</Button>
          </Space>
        </Form>
      </Card>

      <Card
        title="到货计划 - 列表"
        extra={<ListActions loading={loading} onRefresh={reload} onExport={() => exportArrivalPlans(buildQuery(filterForm.getFieldsValue()))} onTemplate={exportArrivalPlanTemplate} onImport={(file) => importFile(file, importArrivalPlans, reload)} />}
      >
        <FilterBar form={filterForm} onSearch={reload} onReset={() => { filterForm.resetFields(); reload(); }} fields={[
          ['q', '关键词'], ['category', '类别'], ['supplier', '供应商'], ['order_no', '订单号'], ['material_code', '物料代码']
        ]} selectFields={{ category: planCategories }} />
        <Table
          rowKey="plan_id"
          loading={loading}
          dataSource={rows}
          pagination={{ pageSize: 10, showTotal: (total) => `共 ${total} 条` }}
          scroll={{ x: 1500 }}
          columns={[
            { title: '类别', dataIndex: 'category', width: 120 },
            { title: '物料代码', dataIndex: 'material_code', width: 160 },
            { title: '物料名称', dataIndex: 'material_name', width: 180 },
            { title: '数量（台）', dataIndex: 'quantity', width: 110 },
            { title: '收货地址', dataIndex: 'receiving_address', width: 220 },
            { title: '供应商', dataIndex: 'supplier', width: 180 },
            { title: '订单号', dataIndex: 'order_no', width: 160 },
            { title: '资产编码区间', dataIndex: 'asset_code_range', width: 180 },
            { title: '预估到货时间', dataIndex: 'estimated_arrival_time', width: 180 },
            { title: '备注', dataIndex: 'remark', width: 180 },
            actionColumn(onEdit, (row) => onDelete(row.plan_id))
          ]}
        />
      </Card>
    </Space>
  );
}

function DeviceArrivalTab() {
  const [form] = Form.useForm<DeviceFormValue>();
  const [filterForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editingID, setEditingID] = useState('');
  const [rows, setRows] = useState<DeviceArrivalRecord[]>([]);

  async function reload() {
    setLoading(true);
    try {
      const resp = ensureApiOk(await listDeviceArrivals(buildQuery(filterForm.getFieldsValue())));
      setRows(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载服务器&网络设备到货记录失败'));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    reload();
  }, []);

  function resetForm() {
    setEditingID('');
    form.resetFields();
  }

  async function onSubmit(values: DeviceFormValue) {
    setSaving(true);
    try {
      const payload = {
        category: values.category,
        package_code: trim(values.package_code),
        package_type: trim(values.package_type),
        material_service_code: trim(values.material_service_code),
        material_service_description: trim(values.material_service_description),
        rack_units: Number(values.rack_units || 0),
        manufacturer: trim(values.manufacturer),
        quantity: Number(values.quantity || 0),
        receiving_location: trim(values.receiving_location),
        purchase_request_no: trim(values.purchase_request_no),
        srm_requirement_submitted_at: formatDateTime(values.srm_requirement_submitted_at),
        po_no: trim(values.po_no),
        actual_arrival_time: formatDateTime(values.actual_arrival_time)
      };
      if (editingID) {
        ensureApiOk(await updateDeviceArrival(editingID, payload));
        message.success('到货记录已更新');
      } else {
        ensureApiOk(await createDeviceArrival(payload));
        message.success('到货记录已创建');
      }
      resetForm();
      await reload();
    } catch (e) {
      message.error(parseApiError(e, editingID ? '更新到货记录失败' : '创建到货记录失败'));
    } finally {
      setSaving(false);
    }
  }

  function onEdit(row: DeviceArrivalRecord) {
    setEditingID(row.record_id);
    form.setFieldsValue({
      ...row,
      srm_requirement_submitted_at: parseDateTime(row.srm_requirement_submitted_at),
      actual_arrival_time: parseDateTime(row.actual_arrival_time)
    });
  }

  async function onDelete(recordId: string) {
    try {
      ensureApiOk(await deleteDeviceArrival(recordId));
      message.success('到货记录已删除');
      if (editingID === recordId) resetForm();
      await reload();
    } catch (e) {
      message.error(parseApiError(e, '删除到货记录失败'));
    }
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card title="服务器&网络设备到货记录 - 新建/编辑" extra={editingID ? <Tag color="blue">编辑中：{editingID}</Tag> : <Tag>新建</Tag>}>
        <Form form={form} layout="vertical" onFinish={onSubmit}>
          <Space wrap>
            <Form.Item name="category" label="类别" rules={[{ required: true, message: '请选择类别' }]} style={{ width: 180 }}>
              <Select options={deviceCategories.map((value) => ({ value, label: value }))} />
            </Form.Item>
            <Form.Item name="package_code" label="套餐代号" rules={[{ required: true, message: '请输入套餐代号' }]} style={{ width: 180 }}>
              <Input />
            </Form.Item>
            <Form.Item name="package_type" label="套餐类型" style={{ width: 180 }}>
              <Input />
            </Form.Item>
            <Form.Item name="material_service_code" label="物料/服务编码" rules={[{ required: true, message: '请输入物料/服务编码' }]} style={{ width: 220 }}>
              <Input />
            </Form.Item>
            <Form.Item name="rack_units" label="U数" style={{ width: 120 }}>
              <InputNumber min={0} precision={2} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="manufacturer" label="厂商" rules={[{ required: true, message: '请输入厂商' }]} style={{ width: 180 }}>
              <Input />
            </Form.Item>
            <Form.Item name="quantity" label="数量" rules={[{ required: true, message: '请输入数量' }]} style={{ width: 120 }}>
              <InputNumber min={1} precision={0} style={{ width: '100%' }} />
            </Form.Item>
          </Space>
          <Space wrap>
            <Form.Item name="receiving_location" label="收货地点" rules={[{ required: true, message: '请输入收货地点' }]} style={{ width: 220 }}>
              <Input />
            </Form.Item>
            <Form.Item name="purchase_request_no" label="采购申请编号" rules={[{ required: true, message: '请输入采购申请编号' }]} style={{ width: 220 }}>
              <Input />
            </Form.Item>
            <Form.Item name="po_no" label="PO号" rules={[{ required: true, message: '请输入PO号' }]} style={{ width: 200 }}>
              <Input />
            </Form.Item>
            <Form.Item name="srm_requirement_submitted_at" label="SRM需求提交时间" rules={[{ required: true, message: '请选择SRM需求提交时间' }]}>
              <DatePicker showTime format={dateTimeFormat} />
            </Form.Item>
            <Form.Item name="actual_arrival_time" label="实际到货时间" rules={[{ required: true, message: '请选择实际到货时间' }]}>
              <DatePicker showTime format={dateTimeFormat} />
            </Form.Item>
          </Space>
          <Form.Item name="material_service_description" label="物料服务描述" style={{ maxWidth: 720 }}>
            <Input.TextArea rows={2} />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={saving}>{editingID ? '保存更新' : '创建记录'}</Button>
            <Button onClick={resetForm} disabled={saving}>重置</Button>
          </Space>
        </Form>
      </Card>

      <Card
        title="服务器&网络设备到货记录 - 列表"
        extra={<ListActions loading={loading} onRefresh={reload} onExport={() => exportDeviceArrivals(buildQuery(filterForm.getFieldsValue()))} onTemplate={exportDeviceArrivalTemplate} onImport={(file) => importFile(file, importDeviceArrivals, reload)} />}
      >
        <FilterBar form={filterForm} onSearch={reload} onReset={() => { filterForm.resetFields(); reload(); }} fields={[
          ['q', '关键词'], ['category', '类别'], ['manufacturer', '厂商'], ['package_code', '套餐代号'], ['purchase_request_no', '采购申请编号'], ['po_no', 'PO号']
        ]} selectFields={{ category: deviceCategories }} />
        <Table
          rowKey="record_id"
          loading={loading}
          dataSource={rows}
          pagination={{ pageSize: 10, showTotal: (total) => `共 ${total} 条` }}
          scroll={{ x: 1800 }}
          columns={[
            { title: '类别', dataIndex: 'category', width: 100 },
            { title: '套餐代号', dataIndex: 'package_code', width: 140 },
            { title: '套餐类型', dataIndex: 'package_type', width: 140 },
            { title: '物料/服务编码', dataIndex: 'material_service_code', width: 170 },
            { title: '物料服务描述', dataIndex: 'material_service_description', width: 220 },
            { title: 'U数', dataIndex: 'rack_units', width: 90 },
            { title: '厂商', dataIndex: 'manufacturer', width: 140 },
            { title: '数量', dataIndex: 'quantity', width: 90 },
            { title: '收货地点', dataIndex: 'receiving_location', width: 150 },
            { title: '采购申请编号', dataIndex: 'purchase_request_no', width: 170 },
            { title: 'SRM需求提交时间', dataIndex: 'srm_requirement_submitted_at', width: 180 },
            { title: 'PO号', dataIndex: 'po_no', width: 140 },
            { title: '实际到货时间', dataIndex: 'actual_arrival_time', width: 180 },
            actionColumn(onEdit, (row) => onDelete(row.record_id))
          ]}
        />
      </Card>
    </Space>
  );
}

function AccessoryArrivalTab() {
  const [form] = Form.useForm<AccessoryFormValue>();
  const [filterForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editingID, setEditingID] = useState('');
  const [rows, setRows] = useState<AccessoryArrivalRecord[]>([]);

  async function reload() {
    setLoading(true);
    try {
      const resp = ensureApiOk(await listAccessoryArrivals(buildQuery(filterForm.getFieldsValue())));
      setRows(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载配件&耗材到货记录失败'));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    reload();
  }, []);

  function resetForm() {
    setEditingID('');
    form.resetFields();
  }

  async function onSubmit(values: AccessoryFormValue) {
    setSaving(true);
    try {
      const payload = {
        purchase_request_no: trim(values.purchase_request_no),
        material_service_code: trim(values.material_service_code),
        material_service_description: trim(values.material_service_description),
        quantity: Number(values.quantity || 0),
        supplier: trim(values.supplier),
        idc_room: trim(values.idc_room),
        purchase_background: trim(values.purchase_background),
        srm_requirement_submitted_at: formatDateTime(values.srm_requirement_submitted_at),
        po_no: trim(values.po_no),
        arrival_time: formatDateTime(values.arrival_time)
      };
      if (editingID) {
        ensureApiOk(await updateAccessoryArrival(editingID, payload));
        message.success('到货记录已更新');
      } else {
        ensureApiOk(await createAccessoryArrival(payload));
        message.success('到货记录已创建');
      }
      resetForm();
      await reload();
    } catch (e) {
      message.error(parseApiError(e, editingID ? '更新到货记录失败' : '创建到货记录失败'));
    } finally {
      setSaving(false);
    }
  }

  function onEdit(row: AccessoryArrivalRecord) {
    setEditingID(row.record_id);
    form.setFieldsValue({
      ...row,
      srm_requirement_submitted_at: parseDateTime(row.srm_requirement_submitted_at),
      arrival_time: parseDateTime(row.arrival_time)
    });
  }

  async function onDelete(recordId: string) {
    try {
      ensureApiOk(await deleteAccessoryArrival(recordId));
      message.success('到货记录已删除');
      if (editingID === recordId) resetForm();
      await reload();
    } catch (e) {
      message.error(parseApiError(e, '删除到货记录失败'));
    }
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card title="配件&耗材到货记录 - 新建/编辑" extra={editingID ? <Tag color="blue">编辑中：{editingID}</Tag> : <Tag>新建</Tag>}>
        <Form form={form} layout="vertical" onFinish={onSubmit}>
          <Space wrap>
            <Form.Item name="purchase_request_no" label="采购申请编号" rules={[{ required: true, message: '请输入采购申请编号' }]} style={{ width: 220 }}>
              <Input />
            </Form.Item>
            <Form.Item name="material_service_code" label="物料/服务代码" rules={[{ required: true, message: '请输入物料/服务代码' }]} style={{ width: 220 }}>
              <Input />
            </Form.Item>
            <Form.Item name="quantity" label="数量" rules={[{ required: true, message: '请输入数量' }]} style={{ width: 120 }}>
              <InputNumber min={1} precision={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="supplier" label="供应商" rules={[{ required: true, message: '请输入供应商' }]} style={{ width: 220 }}>
              <Input />
            </Form.Item>
            <Form.Item name="idc_room" label="机房" rules={[{ required: true, message: '请输入机房' }]} style={{ width: 180 }}>
              <Input />
            </Form.Item>
            <Form.Item name="po_no" label="PO单号" rules={[{ required: true, message: '请输入PO单号' }]} style={{ width: 200 }}>
              <Input />
            </Form.Item>
          </Space>
          <Space wrap>
            <Form.Item name="srm_requirement_submitted_at" label="SRM需求提交时间" rules={[{ required: true, message: '请选择SRM需求提交时间' }]}>
              <DatePicker showTime format={dateTimeFormat} />
            </Form.Item>
            <Form.Item name="arrival_time" label="到货时间" rules={[{ required: true, message: '请选择到货时间' }]}>
              <DatePicker showTime format={dateTimeFormat} />
            </Form.Item>
          </Space>
          <Space wrap align="start">
            <Form.Item name="material_service_description" label="物料/服务描述" style={{ width: 420 }}>
              <Input.TextArea rows={2} />
            </Form.Item>
            <Form.Item name="purchase_background" label="请购背景" style={{ width: 420 }}>
              <Input.TextArea rows={2} />
            </Form.Item>
          </Space>
          <Space>
            <Button type="primary" htmlType="submit" loading={saving}>{editingID ? '保存更新' : '创建记录'}</Button>
            <Button onClick={resetForm} disabled={saving}>重置</Button>
          </Space>
        </Form>
      </Card>

      <Card
        title="配件&耗材到货记录 - 列表"
        extra={<ListActions loading={loading} onRefresh={reload} onExport={() => exportAccessoryArrivals(buildQuery(filterForm.getFieldsValue()))} onTemplate={exportAccessoryArrivalTemplate} onImport={(file) => importFile(file, importAccessoryArrivals, reload)} />}
      >
        <FilterBar form={filterForm} onSearch={reload} onReset={() => { filterForm.resetFields(); reload(); }} fields={[
          ['q', '关键词'], ['supplier', '供应商'], ['idc_room', '机房'], ['purchase_request_no', '采购申请编号'], ['po_no', 'PO单号']
        ]} />
        <Table
          rowKey="record_id"
          loading={loading}
          dataSource={rows}
          pagination={{ pageSize: 10, showTotal: (total) => `共 ${total} 条` }}
          scroll={{ x: 1500 }}
          columns={[
            { title: '采购申请编号', dataIndex: 'purchase_request_no', width: 170 },
            { title: '物料/服务代码', dataIndex: 'material_service_code', width: 170 },
            { title: '物料/服务描述', dataIndex: 'material_service_description', width: 220 },
            { title: '数量', dataIndex: 'quantity', width: 90 },
            { title: '供应商', dataIndex: 'supplier', width: 180 },
            { title: '机房', dataIndex: 'idc_room', width: 140 },
            { title: '请购背景', dataIndex: 'purchase_background', width: 220 },
            { title: 'SRM需求提交时间', dataIndex: 'srm_requirement_submitted_at', width: 180 },
            { title: 'PO单号', dataIndex: 'po_no', width: 140 },
            { title: '到货时间', dataIndex: 'arrival_time', width: 180 },
            actionColumn(onEdit, (row) => onDelete(row.record_id))
          ]}
        />
      </Card>
    </Space>
  );
}

function FilterBar(props: {
  form: ReturnType<typeof Form.useForm>[0];
  fields: Array<[string, string]>;
  selectFields?: Record<string, string[]>;
  onSearch: () => void;
  onReset: () => void;
}) {
  return (
    <Form form={props.form} layout="inline" style={{ marginBottom: 16 }}>
      <Space wrap>
        {props.fields.map(([name, label]) => (
          <Form.Item key={name} name={name} label={label}>
            {props.selectFields?.[name] ? (
              <Select allowClear style={{ width: 150 }} options={props.selectFields[name].map((value) => ({ value, label: value }))} />
            ) : (
              <Input allowClear style={{ width: name === 'q' ? 220 : 160 }} />
            )}
          </Form.Item>
        ))}
        <Button type="primary" onClick={props.onSearch}>查询</Button>
        <Button onClick={props.onReset}>重置</Button>
      </Space>
    </Form>
  );
}

function ListActions(props: {
  loading: boolean;
  onRefresh: () => void;
  onExport: () => void;
  onTemplate: () => void;
  onImport: (file: File) => Promise<void>;
}) {
  const uploadProps: UploadProps = {
    accept: '.xlsx',
    maxCount: 1,
    showUploadList: false,
    customRequest: async (options) => {
      try {
        await props.onImport(options.file as File);
        options.onSuccess?.({}, new XMLHttpRequest());
      } catch {
        options.onError?.(new Error('upload failed'));
      }
    }
  };
  return (
    <Space>
      <Button icon={<DownloadOutlined />} onClick={props.onTemplate}>模板</Button>
      <Upload {...uploadProps}>
        <Button icon={<UploadOutlined />}>导入</Button>
      </Upload>
      <Button icon={<DownloadOutlined />} onClick={props.onExport}>导出</Button>
      <Button onClick={props.onRefresh} loading={props.loading}>刷新</Button>
    </Space>
  );
}

async function importFile(file: File, importer: (file: File) => Promise<any>, reload: () => Promise<void>) {
  try {
    const resp = ensureApiOk(await importer(file));
    const data = resp.data as { created?: number; success?: number; failed?: number };
    const created = data.created ?? data.success ?? 0;
    if ((data.failed || 0) > 0) {
      message.warning(`导入完成，成功 ${created} 条，失败 ${data.failed || 0} 条`);
    } else {
      message.success(`导入成功 ${created} 条`);
    }
    await reload();
  } catch (e) {
    message.error(parseApiError(e, '导入失败'));
    throw e;
  }
}

function actionColumn<T>(onEdit: (row: T) => void, onDelete: (row: T) => void) {
  return {
    title: '操作',
    fixed: 'right' as const,
    width: 150,
    render: (_: unknown, row: T) => (
      <Space>
        <Button size="small" onClick={() => onEdit(row)}>编辑</Button>
        <Popconfirm title="确认删除该记录？" onConfirm={() => onDelete(row)}>
          <Button size="small" danger>删除</Button>
        </Popconfirm>
      </Space>
    )
  };
}

function trim(value?: string) {
  return (value || '').trim();
}

function formatDateTime(value?: Dayjs) {
  return value ? value.format(dateTimeFormat) : '';
}

function parseDateTime(value?: string) {
  return value ? dayjs(value) : undefined;
}

function buildQuery(values: Record<string, any>) {
  const out: Record<string, string> = {};
  Object.entries(values || {}).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') return;
    out[key] = String(value).trim();
  });
  return out;
}
