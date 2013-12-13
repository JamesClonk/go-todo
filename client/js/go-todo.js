// go-todo.js

// =========================================================================
// global objects
// =========================================================================
var login = {
	LoggedIn: false,
	Id: 0,
	AccountId: 0,
	Name: "",
	Email: "",
	Password: "",
	Salt: "",
	Role: "",
	Timestamp: 0,
	Timediff: 0,
	LastAuth: 0
};
var tasks = [];
var sortOrder = {"Priority": false, "Created": false, "LastUpdated": false};
var basePath = "http://localhost:8008";


// =========================================================================
// helper functions
// =========================================================================
function resetLogin() {
	login.LoggedIn = false;
	login.Id = 0;
	login.AccountId = 0;
	login.Name = "";
	login.Email = "";
	login.Password = "";
	login.Salt = "";
	login.Role = "";
	login.Timestamp = 0;
	login.LastAuth = 0;
}

function resetTasks() {
	tasks.length = 0;
	tasks = [];
}

function isLoggedIn() {
	if(login.Id == login.AccountId) {
		showTasks();
		return true;
	} else {
		resetLogin();
		showLogin();
		return false;
	}
}

function showLogin() {
	$("div.login").show();
	$("div.menu").hide();
	$("div.task").hide();
	$("div.newTask").hide();
	$("div.editTask").remove();
}

function showTasks() {
	$("div.login").hide();
	$("div.menu").show();
	$("div.task").show();
	$("div.newTask").hide();
	$("div.editTask").remove();
}

function toggleAddTask() {
	$("div.newTask").toggle();
}

function generateSalt() {
	var i = 0;
	var salt = "";
	while(i < 16){
		var num = (Math.floor((Math.random() * 100)) % 94) + 33;
		if ((num >=33) && (num <=47)) { continue; }
		if ((num >=58) && (num <=64)) { continue; }
		if ((num >=91) && (num <=96)) { continue; }
		if ((num >=123) && (num <=126)) { continue; }
		i++;
		salt += String.fromCharCode(num);
	}
	return salt;
}

function generateHtmlTask(task) {
	var html = '<div class="sixteen columns task" title="Task #'+task.Id+'">' +
					'<div class="four columns taskPriority">' +
						'<div class="time">Last Updated: '+moment.unix(task.LastUpdated).fromNow()+'</div>' +
						'<div class="time">Created: '+moment.unix(task.Created).fromNow()+'</div>' +
					'</div>' +
					'<div class="nine columns taskText">' +
						'<p>'+task.Task+'</p>' +
					'</div>' +
					'<div class="two columns">' +
						'<div class="modify"><a href="#tasks" title="Edit Task #'+task.Id+'">Edit</a></div>' +
						'<div class="modify"><a href="#tasks" title="Remove Task #'+task.Id+'">Remove</a></div>' +
					'</div>' +
				'</div>';
	return html;
}

function addHtmlTask(task) {
	$("div.container").append(generateHtmlTask(task));
}

function updatePriorityGraphic() {
	$("div.task[title^='Task #']").each(function() {

		var taskId = $(this).attr('title').substring(6);
		var task = _.find(tasks, function(task){ return task.Id == taskId; });

		var color;
		switch (task.Priority) {
			case 5:
				color = "#ff3333";
				break;
			case 4:
				color = "#ff9933";
				break;
			case 3:
				color = "#ffff00";
				break;
			case 2:
				color = "#6666ff";
				break;
			case 1:
				color = "#33ff33";
				break;
			default:
		}

		for(i = task.Priority; i > 0; i--) {
			var paper = Raphael($(this).children(".taskPriority")[0], 30, 30);
			paper.circle(15, 15, 14).attr("fill", color);
		}
	});
}

function removeHtmlTasks() {
	$("div.task").remove();
	$("div.editTask").remove();
}

function queryAuth() {
	var timestamp = Math.round((new Date()).getTime() / 1000) - login.Timediff;
	var salt = generateSalt();
	var token = CryptoJS.SHA512(timestamp + salt + login.Password);
	return "?rId=" + login.AccountId + "&rTimestamp=" + timestamp + "&rSalt=" + salt + "&rToken=" + token.toString(CryptoJS.enc.Hex);
}

function sortTasks(field) {
	tasks = _.sortBy(tasks, function(task){ return task[field]; });

	sortOrder[field] = !sortOrder[field];
	if(sortOrder[field]) {
		tasks = tasks.reverse();
	}
}

// =========================================================================
// main functions
// =========================================================================
function loginUser() {
	resetLogin();

	login.Email = $("input[name='email']").val();
	login.Password = $("input[name='password']").val();

	var jsonUrl = basePath + "/auth/?login=" + login.Email;

	// retrieve auth information
	$.getJSON(jsonUrl, function(data) {
		$.each(data, function(key, val) {
	 		login[key] = val;
	 	});

	 	// calculate time difference between local client and server
	 	login.Timediff = Math.round((new Date()).getTime() / 1000) - login.Timestamp;

		var hash = CryptoJS.SHA512(login.Salt + login.Password);
	 	login.Password = hash.toString(CryptoJS.enc.Hex);
	 	
	 	// query /account/ to test password and retrieve more account information
	 	url = basePath + "/account/" + login.AccountId + queryAuth();
	 	$.getJSON(url, function(data) {
		 	$.each(data, function(key, val) {
		 		if(key != "Password") { // don't overwrite the password
		 			login[key] = val;
		 		}
		 	})
		 	loadTasks();

		}).fail(function() {
			resetLogin();
			showLogin();
		});

	}).fail(function() {
		resetLogin();
		showLogin();
	});
}

function loadTasks() {
	if(!isLoggedIn()) { return; }

	// remove all tasks first if there are already some
	resetTasks();
	removeHtmlTasks();

	// retrieve tasks
	var url = basePath + "/tasks/" + queryAuth();
	$.getJSON(url, function(data) {
	 	$.each(data, function(_, task) {;
	 		tasks.push(task);
	 	});
	 	addHtmlTasks(tasks);
	});
}

function addHtmlTasks(tasks) {
	// remove html tasks first
	removeHtmlTasks();
	_.forEach(tasks, addHtmlTask); // add tasks to html
	updatePriorityGraphic();

	$("a[title^='Edit']").click(function() {
		editTask($(this));
	});

	$("a[title^='Remove']").click(function() {
		removeTask($(this));
	});
}

function addNewTask() {
	// send add request
	var data = $('#AddNewTask').serialize();
	var url = basePath + "/task/" + queryAuth();
	$.ajax({
	    url: url,
	    type: 'POST',
	    data: data,
	    datatype: 'json',
	    success: function(result) {
	        if(result["Add"] != "Success") {
	        	alert("ERROR: Could not add new task!");
	        }
	    }
	}).fail(function() {
		alert("ERROR: Could not add new task!");
	});
}

function editTask(element) {
	// only show 1 edit view at max.
	$("div.editTask").remove();
	$("div.task").show();

	var taskId = element.attr('title').substring(11);
	var task = _.find(tasks, function(task){ return task.Id == taskId; });

	taskHtml = element.parent().parent().parent();
	editHtml = 	'<div class="sixteen columns editTask">' +
					'<form id="UpdateTask">' +
						'<input type="hidden" name="Id" value="'+task.Id+'"/>' +
						'<input type="hidden" name="AccountId" value="'+task.AccountId+'"/>' +
						'<div class="three columns">' +
							'Priority:' +
							'<select class="UpdateTask" name="Priority">' +
								'<option value="5">5 - Highest</option>' +
								'<option value="4">4</option>' +
								'<option value="3">3 - Default</option>' +
								'<option value="2">2</option>' +
								'<option value="1">1 - Lowest</option>' +
							'</select> ' +
						'</div>' +
						'<div class="fifteen columns">' +
							'<textarea class="taskText UpdateTask" name="Task">'+task.Task+'</textarea>' +
						'</div>' +
						'<div class="fifteen columns">' +
							'<input type="button" value="Cancel" class="CancelUpdateTask"/>&nbsp;' +
							'<input type="button" value="Update Task" class="UpdateTask"/>' +
						'</div>' +
					'<form>' +
				'</div>';

	taskHtml.before(editHtml);
	$('.UpdateTask[name="Priority"]').val(task.Priority);
	taskHtml.hide();

	$("input.CancelUpdateTask[type='button']").click(function() {
		showTasks();
	});

	$("input.UpdateTask[type='button']").click(function() {
		// send add request
		var data = $('#UpdateTask').serialize();
		var url = basePath + "/task/" + task.Id + queryAuth();
		$.ajax({
		    url: url,
		    type: 'PUT',
		    data: data,
		    datatype: 'json',
		    success: function(result) {
		        if(result["Edit"] != "Success") {
		        	alert("ERROR: Could not edit task!");
		        } else {
			        task.Priority = parseInt($('select.UpdateTask').val());
			        task.Task = $('textarea.UpdateTask').val();
		        	addHtmlTasks(tasks);
		        	showTasks();
			    }
		    }
		}).fail(function() {
			alert("ERROR: Could not edit task!");
			showTasks();
		});
	});
}

function removeTask(element) {
	//[Remove Task #]
	var taskId = element.attr('title').substring(13);
	var task = _.find(tasks, function(task){ return task.Id == taskId; });

	// send deletion request
	var url = basePath + "/task/" + task.Id + queryAuth();
	$.ajax({
	    url: url,
	    type: 'DELETE',
	    datatype: 'json',
	    success: function(result) {
	        if(result["Delete"] != "Success") {
	        	alert("ERROR: Could not remove task!");
	        } else {
	        	loadTasks();
				showTasks();
	        }
	    }
	}).fail(function() {
		alert("ERROR: Could not remove task!");
	});
}