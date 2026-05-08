(function() {
    var video = null;
    var stream = null;
    var scanning = false;
    var canvas = document.getElementById('qr-canvas');
    var ctx = canvas ? canvas.getContext('2d') : null;

    window.handleQRFile = function(input) {
        if (!input.files || !input.files[0]) return;
        stopCamera();
        setStatus('scanning');

        var reader = new FileReader();
        reader.onload = function(e) {
            var img = new Image();
            img.onload = function() {
                canvas.width = img.width;
                canvas.height = img.height;
                ctx.drawImage(img, 0, 0);
                var imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
                var code = jsQR(imageData.data, imageData.width, imageData.height);
                processResult(code);

                var preview = document.getElementById('qr-preview');
                if (preview) {
                    preview.src = e.target.result;
                    preview.style.display = 'block';
                }
            };
            img.src = e.target.result;
        };
        reader.readAsDataURL(input.files[0]);
    };

    window.startCamera = function() {
        var videoEl = document.getElementById('qr-video');
        var preview = document.getElementById('qr-preview');
        if (preview) preview.style.display = 'none';

        if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
            setStatus('camera_error');
            return;
        }

        setStatus('scanning');
        navigator.mediaDevices.getUserMedia({
            video: { facingMode: 'environment', width: { ideal: 640 }, height: { ideal: 480 } }
        }).then(function(s) {
            stream = s;
            video = videoEl;
            video.srcObject = stream;
            video.setAttribute('playsinline', 'true');
            video.play();
            video.style.display = 'block';
            scanning = true;

            document.getElementById('qr-camera-btn').style.display = 'none';
            document.getElementById('qr-camera-stop-btn').style.display = 'inline-flex';

            scanFrame();
        }).catch(function() {
            setStatus('camera_error');
        });
    };

    window.stopCamera = function() {
        scanning = false;
        if (stream) {
            stream.getTracks().forEach(function(t) { t.stop(); });
            stream = null;
        }
        if (video) {
            video.style.display = 'none';
            video = null;
        }
        var camBtn = document.getElementById('qr-camera-btn');
        var stopBtn = document.getElementById('qr-camera-stop-btn');
        if (camBtn) camBtn.style.display = 'inline-flex';
        if (stopBtn) stopBtn.style.display = 'none';
    };

    function scanFrame() {
        if (!scanning || !video || video.readyState !== video.HAVE_ENOUGH_DATA) {
            if (scanning) requestAnimationFrame(scanFrame);
            return;
        }

        canvas.width = video.videoWidth;
        canvas.height = video.videoHeight;
        ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
        var imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
        var code = jsQR(imageData.data, imageData.width, imageData.height);

        if (code) {
            stopCamera();
            processResult(code);
        } else {
            requestAnimationFrame(scanFrame);
        }
    }

    function processResult(code) {
        if (!code) {
            setStatus('not_found');
            return;
        }

        var data = code.data.trim();
        if (!validateLPA(data)) {
            setStatus('invalid');
            document.getElementById('qr-decoded-text').textContent = data;
            document.getElementById('qr-decoded-text').style.display = 'block';
            return;
        }

        var input = document.getElementById('activation_code');
        if (input) {
            input.value = data;
            input.dispatchEvent(new Event('input'));
        }

        setStatus('found');
        document.getElementById('qr-decoded-text').textContent = data;
        document.getElementById('qr-decoded-text').style.display = 'block';
    }

    function validateLPA(s) {
        if (!s) return false;
        var parts = s.split('$');
        if (parts.length < 3) return false;
        if (parts[0] !== 'LPA:1') return false;
        if (!parts[1] || parts[1].indexOf('.') === -1) return false;
        if (!parts[2] || parts[2].length === 0) return false;
        return true;
    }

    function setStatus(status) {
        var el = document.getElementById('qr-status');
        if (!el) return;
        el.className = 'qr-status';
        el.style.display = 'block';

        var msgs = {
            scanning: el.getAttribute('data-msg-scanning'),
            found: el.getAttribute('data-msg-found'),
            not_found: el.getAttribute('data-msg-not-found'),
            invalid: el.getAttribute('data-msg-invalid'),
            camera_error: el.getAttribute('data-msg-camera-error')
        };

        el.textContent = msgs[status] || '';

        if (status === 'found') el.className = 'qr-status qr-status-success';
        else if (status === 'scanning') el.className = 'qr-status qr-status-scanning';
        else el.className = 'qr-status qr-status-error';
    }
})();
