// Hermes eUICC Web — client-side helpers
// Routes all htmx hx-confirm prompts through SweetAlert2 and flags
// destructive actions (delete/remove/memory-reset) with a danger style.

(function () {
    "use strict";

    var i18n = window.SWAL_I18N || { confirm: "OK", cancel: "Cancel" };

    // Paths whose confirmation must use the destructive (red) style.
    var DESTRUCTIVE = /\/(delete|remove|memory-reset)(\/|$)/;

    function requestPath(elt) {
        return (
            elt.getAttribute("hx-post") ||
            elt.getAttribute("hx-get") ||
            elt.getAttribute("hx-put") ||
            elt.getAttribute("hx-delete") ||
            ""
        );
    }

    // Intercept htmx's native confirm and present SweetAlert2 instead.
    document.addEventListener("htmx:confirm", function (evt) {
        var question = evt.detail.question;
        if (!question) {
            return; // No hx-confirm on this element; let htmx proceed.
        }
        evt.preventDefault();

        var danger = DESTRUCTIVE.test(requestPath(evt.detail.elt));

        Swal.fire({
            text: question,
            icon: danger ? "warning" : "question",
            showCancelButton: true,
            reverseButtons: true,
            confirmButtonText: i18n.confirm,
            cancelButtonText: i18n.cancel,
            confirmButtonColor: danger ? "#dc3545" : "#0d6efd",
            cancelButtonColor: "#6c757d"
        }).then(function (result) {
            if (result.isConfirmed) {
                evt.detail.issueRequest(true); // Skip htmx's own confirm.
            }
        });
    });
})();
