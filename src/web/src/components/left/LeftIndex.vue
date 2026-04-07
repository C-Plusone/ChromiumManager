<template>
  <div class="left">
    <div class="header">
      <div class="title"><span>分组列表</span></div>
      <div class="button">
        <span>
          <el-icon @click="onAddClick"><Plus /></el-icon>
        </span>
      </div>
    </div>
    <Content ref="contentRef" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import Content from './LeftContent.vue'

let contentRef = ref(null)
const onAddClick = () => {
  contentRef.value.onAddClick()
}
</script>

<style lang="scss" scoped>
.left {
  float: left;
  width: 280px;
  height: 100%;
  border-right: $border1;

  .header {
    height: 60px;
    line-height: 60px;
    padding: 0 20px;
    background-color: $background-color2;
    color: $white-color;
    font-size: 24px;

    .title {
      float: left;
    }

    .button {
      float: right;

      .el-icon {
        cursor: pointer;
        font-size: 24px;
      }
    }
  }
}
</style>
