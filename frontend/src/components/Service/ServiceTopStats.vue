<template>
    <div class="row stats_area mt-5 mb-4">
        <div class="col-4">
            <span class="font-5 d-block font-weight-bold">{{humanTime(service.avg_response)}}</span>
            <span class="font-1 subtitle">{{$t('average_response')}}</span>
        </div>
        <div class="col-4">
            <span class="font-5 d-block font-weight-bold">{{service.online_24_hours}} %</span>
            <span class="font-1 subtitle">{{$t('last_uptime')}} 24 {{$tc('hour', 24)}}</span>
        </div>
        <div class="col-4">
            <span class="font-5 d-block font-weight-bold">{{service.online_7_days}} %</span>
            <span class="font-1 subtitle">{{$t('last_uptime')}} 7 {{$tc('day', 7)}}</span>
        </div>
    </div>
</template>

<script>
    export default {
        name: 'ServiceTopStats',
        props: {
            service: {
                type: Object,
                required: true
            },
        }
    }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
