window.onload = function () {
    const app = new Vue({
        el: '#app',
        data: {
            pods: [],
            selectedPod: null,
            logs: '',
            connection: null,
        },
        created: function () {
            // Simple GET request using fetch
            fetch("/pods")
                .then(response => response.json())
                .then(data => (this.pods = data));
        },
        methods: {
            selectPod(item){
                this.selectedPod = item
                this.logs = ""
                self = this;
                if (this.connection !== null) {
                    this.connection.close()
                }
                this.connection = new WebSocket(`ws://localhost:8080/namespace/${item.namespace}/pod/${item.pod}/logs`)
                this.connection.onmessage = function(event) {
                    console.log(event)
                    self.logs += `${event.data}<br>`
                }

                this.connection.onopen = function(event) {
                    console.log(event)
                    console.log("Successfully connected to the echo websocket server...")
                }
            }
        }
    })
}

