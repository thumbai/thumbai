function markFieldError(err) {
    $('#'+err.Name).addClass('is-invalid');
    $('#'+err.Name+'Error').html(err.Message);
}

function markFieldErrors(errs) {
    for (var i in errs) {
        markFieldError(errs[i]);
    }
}

function applyOnKeydownRemoveError(formId) {
    $('#'+formId).find('.form-control').each(function(){
        $(this).keydown(function(){
            $(this).removeClass('is-invalid');
        });
    })
}

function disableWithSpinner(id) {
    var obj = $('#'+id);
    var text = obj.html();
    obj.data('previous-text', text);
    obj.prop('disabled', true).html('<i class="fas fa-spinner fa-pulse pr-1"></i> ' + text);
}

function enableWithoutSpinner(id) {
    var obj = $('#'+id);
    obj.prop('disabled', false).html(obj.data('previous-text'));
}

function showConfirmYesNo(confirmText, callback) {
    $('#confirmText').html(confirmText);
    $('#confirmDialog').show();
}

$.extend({ confirmDialog: function (confirmText, confirmTarget, yesCallback) {
    $('<div class="modal fade" id="confirmDialog" tabindex="-1" role="dialog" aria-labelledby="addEditModalTitle" aria-hidden="true">'+
    '<div class="modal-dialog modal-dialog-centered" role="document">' +
        '<div class="modal-content pr-2 pl-2">' +
            '<div class="modal-body">'+
                '<div class="p-1 mt-2">'+
                    '<div id="confimDialogText"></div>'+
                '</div>' +
                '<div class="mt-1 mb-5">' +
                    '<div class="float-right">' +
                        '<button type="button" class="no btn btn-sm btn-outline-secondary pl-3 pr-3 mr-1" data-dismiss="modal">No</button>'+
                        '<button type="button" id="configmDialogYesBtn" class="yes btn btn-sm btn-danger pl-3 pr-3">Yes</button>' +
                    '</div>' +
                '</div>'+
            '</div>'+
        '</div>'+
    '</div>'+
'</div>').appendTo('body');
    $('#confimDialogText').html(confirmText);
    $('#configmDialogYesBtn').click(function(){
        $('#confirmDialog').modal('hide');
        yesCallback(confirmTarget);
    });
    $('#confirmDialog').modal('show');
}});

function csrfSafeMethod(method) {
    // these HTTP methods do not require Anti-CSRF
    return (/^(GET|HEAD|OPTIONS|TRACE)$/.test(method));
}

function antiCsrfHeader() {
    if (!this.crossDomain) {
        var antiCSRFToken = $('meta[name="anti_csrf_token"]').attr('content');
        return {'X-Anti-CSRF-Token': antiCSRFToken}
    }
    return {}
}

function showFeedback(mode, text) {
    var cssClass = mode === 'success' ? 'text-success' : 'text-danger';
    $('#genericFeedback').html('<strong class="'+ cssClass +'">'+text+'</strong>').fadeIn();
}

function fadeOutFeedback(delay) {
    setTimeout(function () { 
        $('#genericFeedback').fadeOut('slow', function(){
            $(this).removeClass('text-danger text-success');
        });
    }, delay);
}

function showFormFeedback(mode, text) {
    var cssClass = mode === 'success' ? 'text-success' : 'text-danger';
    $('#formFeedback').html('<strong class="'+ cssClass +'">'+text+'</strong>').fadeIn();
}

function fadeOutFormFeedback(delay) {
    setTimeout(function () { 
        $('#formFeedback').fadeOut('slow', function(){
            $(this).removeClass('text-danger text-success');
        }); 
    }, delay);
}