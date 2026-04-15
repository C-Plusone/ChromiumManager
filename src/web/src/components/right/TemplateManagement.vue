<template>
  <el-dialog v-model="visible" title="模板列表" :width="900" :before-close="handleClose">
    <div class="tpl-header">
      <el-input
        v-model="keyword"
        placeholder="搜索模板名称"
        clearable
        class="tpl-search"
        @keydown.enter="onSearchClick"
        @clear="onSearchClick"
      >
        <template #append>
          <el-button :icon="Search" @click="onSearchClick"></el-button>
        </template>
      </el-input>
      <el-button type="primary" :icon="Plus" @click="onAddClick">添加模板</el-button>
    </div>

    <div class="tpl-table-wrap">
      <el-table
        v-loading="loading"
        :data="list"
        class="tpl-table"
        highlight-current-row
        stripe
      >
        <el-table-column prop="name" label="模板名称" width="auto" show-overflow-tooltip></el-table-column>
        <el-table-column label="代理" width="120" show-overflow-tooltip>
          <template #default="scope">{{ getProxyName(scope.row.proxy) }}</template>
        </el-table-column>
        <el-table-column label="随机指纹" width="80" align="center">
          <template #default="scope">
            <el-tag v-if="scope.row.fingerprint?.randomFingerprint" type="success" size="small">是</el-tag>
            <el-tag v-else type="info" size="small">否</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作系统" width="80" show-overflow-tooltip>
          <template #default="scope">{{ scope.row.fingerprint?.platform || '-' }}</template>
        </el-table-column>
        <el-table-column fixed="right" label="操作" width="220">
          <template #default="scope">
            <el-button type="primary" size="small" @click.stop="onUseClick(scope.row)">
              使用
            </el-button>
            <el-button size="small" @click.stop="onEditClick(scope.row)">编辑</el-button>
            <el-button size="small" class="delete" @click.stop="onDeleteClick(scope.row)">
              删除
            </el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无模板"></el-empty>
        </template>
      </el-table>
    </div>

    <el-pagination
      v-model:current-page="page.current"
      v-model:page-size="page.size"
      :total="page.total"
      layout="total, prev, pager, next"
      class="tpl-pagination"
      @current-change="handleCurrentChange"
    ></el-pagination>

    <!-- Form for Add/Edit Template -->
    <el-dialog
      v-model="formDialog"
      :title="form._id ? '编辑模板' : '添加模板'"
      :width="800"
      append-to-body
      :before-close="onFormCancel"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-position="right" label-width="auto" size="default">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="模板名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入模板名称"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="代理" class="input-picker">
              <el-input
                :model-value="formProxyName"
                placeholder="请选择代理"
                @click="onManageProxyClick"
                @keydown.prevent
              >
                <template #suffix>
                  <span
                    class="input-picker-suffix"
                    @click.stop="form.proxy ? (form.proxy = '') : onManageProxyClick()"
                  >
                    <el-icon v-if="form.proxy" class="el-input__clear"><CircleClose /></el-icon>
                    <el-icon v-else><ArrowDown /></el-icon>
                  </span>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row class="switches-row">
          <el-col>
            <el-form-item label="随机指纹">
              <el-switch v-model="form.fp.randomFingerprint" />
            </el-form-item>
          </el-col>
          <el-col>
            <el-form-item label="代理语言">
              <el-switch v-model="form.fp.proxyLang" />
            </el-form-item>
          </el-col>
          <el-col>
            <el-form-item label="代理时区">
              <el-switch v-model="form.fp.proxyTimezone" />
            </el-form-item>
          </el-col>
          <el-col>
            <el-form-item label="代理位置">
              <el-switch v-model="form.fp.proxyLocation" />
            </el-form-item>
          </el-col>
          <el-col>
            <el-form-item label="WebRTC">
              <el-switch
                :model-value="!form.fp.disableFeatures.includes('webrtc')"
                @update:model-value="(v) => toggleFeature('webrtc', !v)"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="操作系统">
              <el-select v-model="form.fp.platform" clearable placeholder="请选择操作系统">
                <el-option label="Windows" value="windows"></el-option>
                <el-option label="Linux" value="linux"></el-option>
                <el-option label="macOS" value="macos"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="浏览器品牌">
              <el-select v-model="form.fp.brand" clearable placeholder="请选择浏览器品牌">
                <el-option label="Chrome" value="Chrome"></el-option>
                <el-option label="Edge" value="Edge"></el-option>
                <el-option label="Opera" value="Opera"></el-option>
                <el-option label="Vivaldi" value="Vivaldi"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="设备核心">
              <el-select v-model="form.fp.hardwareConcurrency" placeholder="请选择设备核心" clearable>
                <el-option label="2" value="2"></el-option>
                <el-option label="4" value="4"></el-option>
                <el-option label="6" value="6"></el-option>
                <el-option label="8" value="8"></el-option>
                <el-option label="10" value="10"></el-option>
                <el-option label="12" value="12"></el-option>
                <el-option label="16" value="16"></el-option>
                <el-option label="20" value="20"></el-option>
                <el-option label="24" value="24"></el-option>
                <el-option label="32" value="32"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="设备内存">
              <el-select v-model="form.fp.deviceMemory" placeholder="请选择设备内存" clearable>
                <el-option label="2" value="2"></el-option>
                <el-option label="4" value="4"></el-option>
                <el-option label="8" value="8"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="屏幕尺寸">
              <el-select v-model="form.fp.screen" clearable placeholder="请选择屏幕尺寸">
                <el-option v-for="s in screens" :key="s" :label="s" :value="s"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="位置" class="input-picker">
              <el-input
                v-model="form.fp.location"
                :disabled="form.fp.proxyLocation"
                placeholder="请选择位置"
                clearable
                @click="onPickFpLocation"
                @keydown.prevent
              ></el-input>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="语言">
              <el-select v-model="form.fp.lang" filterable clearable :disabled="form.fp.proxyLang" placeholder="请选择语言">
                <el-option v-for="lang in languages" :key="lang" :label="lang" :value="lang"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="时区">
              <el-select v-model="form.fp.timezone" filterable clearable :disabled="form.fp.proxyTimezone" :fit-input-width="true" placeholder="请选择时区">
                <el-option v-for="tz in timezones" :key="tz" :label="tz" :value="tz"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="禁用伪装">
              <el-checkbox-group v-model="form.fp.disableFingerprint" :disabled="!form.fp.randomFingerprint">
                <el-checkbox value="font">字体</el-checkbox>
                <el-checkbox value="audio">音频</el-checkbox>
                <el-checkbox value="canvas">Canvas</el-checkbox>
                <el-checkbox value="clientrects">ClientRects</el-checkbox>
                <el-checkbox value="webgl">WebGL</el-checkbox>
                <el-checkbox value="gpu">GPU</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="额外参数">
              <el-input
                v-model="form.args"
                placeholder="请输入额外参数，多个以空格分隔"
                type="textarea"
                :rows="3"
                resize="none"
              ></el-input>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <template #footer>
        <el-button @click="onFormCancel">取消</el-button>
        <el-button type="primary" @click="onFormConfirm">确定</el-button>
      </template>
    </el-dialog>

    <!-- Use Template: quick create dialog -->
    <el-dialog
      v-model="useDialog"
      title="从模板创建配置"
      :width="500"
      append-to-body
      :before-close="() => (useDialog = false)"
    >
      <el-form ref="useFormRef" :model="useForm" :rules="useRules" label-width="auto">
        <el-form-item label="模板" prop="templateName">
          <el-input :model-value="useForm.templateName" disabled></el-input>
        </el-form-item>
        <el-form-item label="配置名称" prop="name">
          <el-input v-model="useForm.name" placeholder="请输入新配置名称"></el-input>
        </el-form-item>
        <el-form-item label="分组" prop="groupId">
          <el-select v-model="useForm.groupId" filterable placeholder="请选择分组">
            <template v-for="item in groups" :key="item._id">
              <el-option v-if="item._id !== 'all'" :label="item.name" :value="item._id"></el-option>
            </template>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="useDialog = false">取消</el-button>
        <el-button type="primary" @click="onUseConfirm">创建</el-button>
      </template>
    </el-dialog>

    <ProxyManagement v-model="proxyManageVisible" @select="onProxySelect" />
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch, nextTick, inject, computed } from 'vue'
import { Plus, Search, ArrowDown, CircleClose } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTemplates, getTemplate, addTemplate, updateTemplate, deleteTemplate, createFromTemplate } from '@/api'
import { languages, timezones, screens } from '@/utils/constants'
import { openMapPicker } from '@/utils/mapPicker'
import ProxyManagement from './ProxyManagement.vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue', 'change'])

const device = inject('device')
const proxies = inject('proxies')
const fetchProxies = inject('fetchProxies')

const groups = computed(() => device.group)
const proxyMap = computed(() => new Map(proxies.value.map((p) => [p._id, p])))
const getProxyName = (id) => proxyMap.value.get(id)?.name || ''

const visible = ref(false)
const loading = ref(false)
const list = ref([])
const page = reactive({ current: 1, size: 8, total: 0 })
const keyword = ref('')

const formDialog = ref(false)
const formRef = ref(null)
const form = reactive({ fp: defaultFp() })
const rules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }]
}
const proxyManageVisible = ref(false)

const useDialog = ref(false)
const useFormRef = ref(null)
const useForm = reactive({ templateId: '', templateName: '', name: '', groupId: '' })
const useRules = {
  name: [{ required: true, message: '请输入配置名称', trigger: 'blur' }],
  groupId: [{ required: true, message: '请选择分组', trigger: 'change' }]
}

function defaultFp() {
  return {
    platform: '',
    brand: '',
    hardwareConcurrency: '',
    deviceMemory: '',
    disableFeatures: ['webrtc'],
    screen: '',
    lang: '',
    timezone: '',
    location: '',
    disableFingerprint: [],
    randomFingerprint: true,
    proxyLang: true,
    proxyTimezone: true,
    proxyLocation: true
  }
}

const formProxyName = computed(() => {
  const p = proxyMap.value.get(form.proxy)
  return p ? p.name : ''
})

watch(
  () => props.modelValue,
  (val) => {
    visible.value = val
    if (val) {
      keyword.value = ''
      page.current = 1
      fetchList()
    }
  }
)

const handleClose = () => emit('update:modelValue', false)

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getTemplates({ page: page.current, pageSize: page.size, keyword: keyword.value })
    if (res) {
      list.value = res.list || []
      page.total = res.total || 0
    }
  } catch (err) {
    console.error(err)
  }
  loading.value = false
}

const handleCurrentChange = (val) => {
  page.current = val
  fetchList()
}

const onSearchClick = () => {
  page.current = 1
  fetchList()
}

const onAddClick = () => {
  Object.assign(form, { _id: undefined, name: '', sort: 0, proxy: '', args: '', notes: '', fp: defaultFp() })
  formDialog.value = true
  nextTick(() => formRef.value?.clearValidate())
}

const onEditClick = async (row) => {
  try {
    const data = await getTemplate(row._id)
    let fp = defaultFp()
    if (data.fingerprint && typeof data.fingerprint === 'object') {
      Object.assign(fp, data.fingerprint)
    }
    Object.assign(form, { ...data, fp })
    formDialog.value = true
    nextTick(() => formRef.value?.clearValidate())
  } catch (err) {
    if (!err?.silent) ElMessage.error('获取失败: ' + (err?.message || err))
  }
}

const onFormCancel = () => { formDialog.value = false }

const buildPayload = () => {
  const { fp, ...rest } = form
  return { ...rest, fingerprint: fp }
}

const onFormConfirm = async () => {
  if (!formRef.value) return
  const valid = await new Promise((resolve) => formRef.value.validate(resolve))
  if (!valid) return
  try {
    if (form._id) {
      await updateTemplate(buildPayload())
      ElMessage.success('编辑成功')
    } else {
      await addTemplate(buildPayload())
      ElMessage.success('添加成功')
    }
    formDialog.value = false
    fetchList()
  } catch (err) {
    if (!err?.silent) ElMessage.error((form._id ? '编辑失败：' : '添加失败：') + (err?.message || err))
  }
}

const onDeleteClick = (row) => {
  ElMessageBox.confirm('是否删除该模板？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteTemplate(row._id)
      ElMessage.success('删除成功')
      fetchList()
    } catch (err) {
      if (!err?.silent) ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

const onUseClick = (row) => {
  Object.assign(useForm, { templateId: row._id, templateName: row.name, name: '', groupId: '' })
  useDialog.value = true
  nextTick(() => useFormRef.value?.clearValidate())
}

const onUseConfirm = async () => {
  if (!useFormRef.value) return
  const valid = await new Promise((resolve) => useFormRef.value.validate(resolve))
  if (!valid) return
  try {
    await createFromTemplate({
      templateId: useForm.templateId,
      name: useForm.name,
      groupId: useForm.groupId
    })
    ElMessage.success('创建成功！')
    useDialog.value = false
    emit('change')
  } catch (err) {
    if (!err?.silent) ElMessage.error('创建失败：' + (err?.message || err))
  }
}

const toggleFeature = (feature, enabled) => {
  const arr = form.fp.disableFeatures
  if (enabled) {
    if (!arr.includes(feature)) arr.push(feature)
  } else {
    const idx = arr.indexOf(feature)
    if (idx > -1) arr.splice(idx, 1)
  }
}

const onPickFpLocation = async () => {
  if (form.fp.proxyLocation) return
  const loc = await openMapPicker(form.fp.location || '')
  if (loc) form.fp.location = loc
}

const onManageProxyClick = () => { proxyManageVisible.value = true }
const onProxySelect = async (proxyId) => {
  await fetchProxies()
  form.proxy = proxyId
}
</script>

<style lang="scss">
.tpl-table {
  .el-table__body-wrapper .el-scrollbar__view {
    height: 100%;
  }
}
</style>

<style lang="scss" scoped>
.tpl-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.tpl-search {
  width: 260px;
}
.tpl-table-wrap {
  height: 386px;
}
.tpl-table {
  width: 100%;
  height: 100%;
  padding-top: 15px;
}
.tpl-pagination {
  padding: 15px 0 20px;
  justify-content: center;
}
.delete {
  color: $red-color;
}
.el-select {
  width: 100%;
}
.switches-row {
  :deep(.el-col) {
    flex: 1;
  }
}
</style>
