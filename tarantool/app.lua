box.cfg{listen=3301}

dialogs = box.schema.create_space('dialogs')

dialogs:format({
    { name = 'id', type = 'string' },
    { name = 'author_id', type = 'string' },
    { name = 'recepient_id', type = 'string' },
    { name = 'dialog_id', type = 'string' },
    { name = 'created_at', type = 'unsigned' },
    { name = 'text', type = 'string' },

})

dialogs:create_index(
    'primary', { parts={ {'id'}} }
)

dialogs:create_index('dialog', { parts = { { 'dialog_id' } }, unique = false })

uuid = require('uuid')

function send_message(authorID, recepientID, dialogID, text)
    box.begin()
    dialogs:insert{uuid.str(), authorID, recepientID, dialogID, os.time(os.date("!*t")), text}
    box.commit()
end

function get_dialog(dialog_id)
    box.begin()
    x = dialogs.index.dialog:select{dialog_id}
    box.commit()
    return x
end
