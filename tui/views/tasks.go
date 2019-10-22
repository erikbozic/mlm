package views

import (
	"fmt"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/rivo/tview"
	"log"
	"sort"
)

type TasksView struct {
	*tview.Table
	 taskNames []string
}

func (v *TasksView) SetTasks(tasks map[string][]mesos.Task) {
	i := 0
	v.Clear()
	v.setHeaders()
	v.taskNames = make([]string, len(tasks)) // TODO possible RC
	for taskName, _ := range tasks {
		v.taskNames[i] = taskName
		i++
	}
	sort.Strings(v.taskNames) // sort because we can't guarantee the map will be sorted in any way
	for i, name := range v.taskNames {
		nameCell := tview.NewTableCell(name)
		running := 0
		for _, task := range tasks[name] {
			if *task.State == mesos.TASK_RUNNING {
				running++
			}
 		}
		vCell := tview.NewTableCell(fmt.Sprintf("(%d)", running,))
		v.SetCell(i+1, 0, nameCell)
		v.SetCell(i+1, 1, vCell)
		i++
	}
}
func NewTasksViews(ch chan interface{}) *TasksView {
	v := &TasksView{
		Table: tview.NewTable().Select(1, 1).SetFixed(1, 1).SetSelectable(true, false),

	}


	v.SetSelectedFunc(func(r int, c int){
		cell := v.GetCell(r, 0)
		log.Printf("selected row:%d, col:%d, with text: %s\n", r, c, cell.Text)
		ch <- cell.Text
	})

	v.SetTitle("tasks")
	v.SetBorder(true)
	return v
}

func (v *TasksView) setHeaders(){
	taskNameHeader := tview.NewTableCell("task name").SetExpansion(1).SetSelectable(false)
	instancesHeader := tview.NewTableCell("instances").SetExpansion(5).SetSelectable(false)
	v.SetCell(0,0, taskNameHeader)
	v.SetCell(0,1, instancesHeader)
}






