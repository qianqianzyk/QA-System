package adminService

import (
	"QA-System/app/models"
	"QA-System/config/config"
	"QA-System/config/database"
	"os"
	"strings"
	"time"
)

type Option struct {
	SerialNum  int    `json:"serial_num"`  //选项序号
	Content    string `json:"content"`     //选项内容
	OptionType int    `json:"option_type"` //选项类型 1文字2图片
}

type Question struct {
	ID           int      `json:"id"`
	SerialNum    int      `json:"serial_num"`    //题目序号
	Subject      string   `json:"subject"`       //问题
	Description  string   `json:"description"`   //问题描述
	Img          string   `json:"img"`           //图片
	Required     bool     `json:"required"`      //是否必填
	Unique       bool     `json:"unique"`        //是否唯一
	QuestionType int      `json:"question_type"` //问题类型 1单选2多选3填空4简答5图片
	Reg          string   `json:"reg"`           //正则表达式
	Options      []Option `json:"options"`       //选项
}

func GetSurveyByID(id int) (models.Survey, error) {
	var survey models.Survey
	err := database.DB.Where("id = ?", id).First(&survey).Error
	return survey, err
}

func CreateSurvey(id int,title string, desc string, img string, questions []Question, status int, time time.Time) error {
	var survey models.Survey
	survey.UserID = id
	survey.Title = title
	survey.Desc = desc
	survey.Img = img
	survey.Status = status
	survey.Deadline = time

	err := database.DB.Create(&survey).Error
	if err != nil {
		return err
	}
	for _, question := range questions {
		var q models.Question
		q.SurveyID = survey.ID
		q.SerialNum = question.SerialNum
		q.Subject = question.Subject
		q.Description = question.Description
		q.Img = question.Img
		q.Required = question.Required
		q.Unique = question.Unique
		q.QuestionType = question.QuestionType
		err := database.DB.Create(&q).Error
		if err != nil {
			return err
		}
		for _, option := range question.Options {
			var o models.Option
			o.QuestionID = q.ID
			o.Content = option.Content
			o.SerialNum = option.SerialNum
			o.OptionType = option.OptionType
			err := database.DB.Create(&o).Error
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func UpdateSurveyStatus(id int, status int) error {
	var survey models.Survey
	err := database.DB.Model(&survey).Where("id = ?", id).Update("status", status).Error
	return err
}

func UpdateSurvey(id int, title string, desc string, img string, questions []Question, time time.Time) error {
	//遍历原有问题，删除对应选项
	var survey models.Survey
	var oldQuestions []models.Question
	var old_imgs []string
	var new_imgs []string
	err := database.DB.Where("survey_id = ?", id).Find(&oldQuestions).Error
	if err != nil {
		return err
	}
	old_imgs,err= getOldImgs(id,oldQuestions)
	if err != nil {
		return err
	}
	for _, oldQuestion := range oldQuestions {
		err = database.DB.Where("question_id = ?", oldQuestion.ID).Delete(&models.Option{}).Error
		if err != nil {
			return err
		}
	}
	err = database.DB.Where("survey_id = ?", id).Delete(&models.Question{}).Error
	if err != nil {
		return err
	}
	//修改问卷信息
	err = database.DB.Model(&survey).Where("id = ?", id).Updates(map[string]interface{}{"title": title, "desc": desc, "img": img, "deadline": time}).Error
	if err != nil {
		return err
	}
	new_imgs = append(new_imgs, img)
	//重新添加问题和选项
	for _, question := range questions {
		var q models.Question
		q.SurveyID = survey.ID
		q.Subject = question.Subject
		q.Description = question.Description
		q.Img = question.Img
		q.Required = question.Required
		q.Unique = question.Unique
		q.QuestionType = question.QuestionType
		new_imgs = append(new_imgs, question.Img)
		err := database.DB.Create(&q).Error
		if err != nil {
			return err
		}
		for _, option := range question.Options {
			var o models.Option
			o.QuestionID = q.ID
			o.Content = option.Content
			o.SerialNum = option.SerialNum
			o.OptionType = option.OptionType
			if option.OptionType == 2 {
				new_imgs = append(new_imgs, option.Content)
			}
			err := database.DB.Create(&o).Error
			if err != nil {
				return err
			}
		}
	}
	urlHost := config.Config.GetString("url.host")
	//删除无用图片
	for _, old_img := range old_imgs {
		if !contains(new_imgs, old_img) {
			_ = os.Remove("./static/" + strings.TrimPrefix(old_img, urlHost+"/static/"))
		}
	}
	return nil
}

func UserInManage(uid int, sid int) bool {
	var survey models.Manage
	err := database.DB.Where("user_id = ? and survey_id = ?", uid, sid).First(&survey).Error
	return err == nil
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func getOldImgs(id int, questions []models.Question) ([]string, error) {
	var imgs []string
	var survey models.Survey
	err := database.DB.Where("id = ?", id).First(&survey).Error
	if err != nil {
		return nil, err
	}
	imgs = append(imgs, survey.Img)
	for _, question := range questions {
		imgs = append(imgs, question.Img)
		var options []models.Option
		err = database.DB.Where("question_id = ?", question.ID).Find(&options).Error
		if err != nil {
			return nil, err
		}
		for _, option := range options {
			if option.OptionType == 2 {
				imgs = append(imgs, option.Content)
			}
		}
	}
	return imgs, nil
}