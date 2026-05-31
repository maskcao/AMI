<script setup lang="ts"></script>

<template>
  <el-config-provider>
    <router-view />
  </el-config-provider>
</template>

<style>
#app {
  height: 100%;
  background: var(--ami-bg);
}
</style>
