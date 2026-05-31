<script setup lang="ts">
defineProps<{
  title: string
  total: number
  page: number
  pageSize: number
}>()

const emit = defineEmits<{
  pageChange: [page: number]
  sizeChange: [size: number]
}>()
</script>

<template>
  <el-card shadow="never" class="entity-list-panel">
    <template #header>
      <div class="panel-head">
        <span>{{ title }}</span>
        <span class="panel-total">共 {{ total }} 条</span>
      </div>
    </template>

    <div class="list-body" :class="{ 'has-aside': !!$slots.aside }">
      <aside v-if="$slots.aside" class="list-aside">
        <slot name="aside" />
      </aside>
      <div class="list-main">
        <slot />
        <div v-if="total > 0" class="list-pagination">
          <el-pagination
            :current-page="page"
            :page-size="pageSize"
            :total="total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            background
            small
            @current-change="emit('pageChange', $event)"
            @size-change="emit('sizeChange', $event)"
          />
        </div>
        <el-empty v-else description="无匹配数据" :image-size="72" />
      </div>
    </div>
  </el-card>
</template>

<style scoped>
.entity-list-panel {
  width: 100%;
  background: var(--ami-surface) !important;
}
.list-main :deep(.el-table) {
  width: 100%;
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.panel-total {
  font-size: 13px;
  font-weight: 400;
  color: var(--ami-text-muted);
}
.list-body {
  display: block;
}
.list-body.has-aside {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}
.list-aside {
  position: sticky;
  top: 12px;
}
.list-main {
  min-width: 0;
}
.list-pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
