import axios from 'axios';
import type { ApiResp } from './types';

const codeMessageMap: Record<number, string> = {
  40001: '请求参数有误，请检查后重试',
  40002: '文件格式不支持，请上传 .xlsx',
  40003: '文件读取失败，请确认文件内容后重试',
  40004: '导入字段不匹配，请检查列名',
  40401: '记录不存在，请确认 ID',
  50001: '服务暂时不可用，请稍后重试'
};

export class BizError extends Error {
  constructor(
    public code: number,
    message: string,
    public requestId?: string
  ) {
    super(message);
    this.name = 'BizError';
  }
}

export function ensureApiOk<T>(resp: ApiResp<T>): ApiResp<T> {
  if (resp.code === 0) return resp;
  const requestId = (resp.data as { request_id?: string } | null | undefined)?.request_id;
  throw new BizError(resp.code, resp.message || codeMessageMap[resp.code] || '操作失败', requestId);
}

export function parseApiError(err: unknown, fallback = '操作失败'): string {
  if (axios.isAxiosError(err)) {
    const payload = err.response?.data as { code?: number; message?: string; data?: { request_id?: string } } | undefined;
    const code = payload?.code;
    const serverMessage = payload?.message?.trim();
    const friendly = code ? codeMessageMap[code] : '';
    const requestId = payload?.data?.request_id;

    let base = serverMessage || friendly || err.message || fallback;
    if (friendly && serverMessage && !serverMessage.includes(friendly)) {
      base = `${friendly}（${serverMessage}）`;
    }
    if (requestId) {
      base += ` [请求ID: ${requestId}]`;
    }
    return base;
  }

  if (err instanceof BizError) {
    const friendly = codeMessageMap[err.code];
    let msg = err.message || friendly || fallback;
    if (friendly && err.message && !err.message.includes(friendly)) {
      msg = `${friendly}（${err.message}）`;
    }
    if (err.requestId) msg += ` [请求ID: ${err.requestId}]`;
    return msg;
  }

  if (err instanceof Error && err.message) return err.message;
  return fallback;
}
