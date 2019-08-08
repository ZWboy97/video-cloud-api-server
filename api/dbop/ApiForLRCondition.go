package dbop

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"go-api-server/api/defs"
	"go-api-server/api/utils"
	"log"
)

func InsertLRConditionByCom(lid, verificationCode string, condition ,conditionType, duration, tryToSee int, price float32) (*defs.LiveRoomCondition, error) {

	stmtIns, err := dbConn.Prepare("INSERT INTO live_condition (lid, lcondition, condition_type, price, duration, try_to_see, verification_code) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("pareparation:%v", err)
		return nil, err
	}

	_, err = stmtIns.Exec(lid, condition, conditionType, price, duration, tryToSee, verificationCode)
	if err != nil {
		return nil, err
	}

	log.Printf(" Insert success")

	defer stmtIns.Close()

	LRCON := &defs.LiveRoomCondition{}
	LRCON.Lid = lid
	LRCON.Condition = condition
	LRCON.ConditionType = conditionType
	LRCON.Price = price
	LRCON.Duration = duration
	LRCON.TryToSee = tryToSee
	LRCON.VerificationCode = verificationCode

	return LRCON, nil
}

func UpdateLRConditionByLid(lid, verificationCode, email string, condition, conditionType, duration, tryToSee int, price float32) (*defs.LiveRoomCondition, error) {
	Condition := defs.LiveRoomDefaultConfig
	if condition == 0 { //表示此时观看条件为无条件观看
		stmtUpa, err := dbConn.Prepare("UPDATE live_condition SET lcondition = ?, condition_type = ?, price = ?, duration = ?, try_to_see = ?, verification_code = ? WHERE lid = ?")
		if err != nil {
			log.Printf("Error of preparation of update live_condition_1:%v", err)
			return nil, err
		}
		_, err = stmtUpa.Exec(condition, Condition.ConditionType, Condition.Price, Condition.Duration, Condition.TryToSee, Condition.VerificationCode,lid)
		if err != nil {
			log.Printf("Error of execution of update live_condition_1:%v", err)
			return nil,err
		}
		log.Printf("Update live_condition_1 success\n")
		defer stmtUpa.Close()
		roomCondition := &defs.LiveRoomCondition{}
		roomCondition.Condition = 0
		return roomCondition, nil
	} else {
		if conditionType == 1 { //表示观看条件为付费观看
			stmtUpa, err := dbConn.Prepare("UPDATE live_condition SET lcondition = 1, condition_type = 1, price = ?, duration = ?, try_to_see = ?, verification_code = ? WHERE lid = ?")
			if err != nil {
				log.Printf("Error of preparation of update live_condition_2:%v", err)
				return nil, err
			}
			_, err = stmtUpa.Exec(price, duration, tryToSee, Condition.VerificationCode, lid)
			if err != nil {
				log.Printf("Error of execution of update live_condition_2:%v", err)
				return nil,err
			}
			log.Printf("Update live_condition_2 success\n")
			defer stmtUpa.Close()
			roomCondition := &defs.LiveRoomCondition{}
			roomCondition.Condition = 1
			roomCondition.ConditionType = 1
			roomCondition.TryToSee = tryToSee
			roomCondition.Duration = duration
			roomCondition.Price = price
			return roomCondition, nil

		} else if conditionType == 2 { //此种观看条件为白名单观看
			if email != "" {
				if ok, err := IsEmailRegister(email); !(!ok && err == nil) { //表示邮箱未注册的情况
					log.Printf("邮箱未注册")
					return nil, sql.ErrNoRows
				}
				stmtIns, err := dbConn.Prepare("INSERT INTO whitelist(lid, email) VALUES (?, ?)")
				if err != nil {
					log.Printf("Error of preparation of insert whitelist:%v", err)
					return nil,err
				}
				_, err = stmtIns.Exec(lid, email)
				if err != nil {
					log.Printf("Error of insertion whitelist:%v", err)
					return nil, err
				}


				stmtUpa, err := dbConn.Prepare("UPDATE live_condition SET lcondition = 1, condition_type = 2, price = ?, duration = ?, try_to_see = ?, verification_code = ? WHERE lid = ?")
				if err != nil {
					log.Printf("Error of preparation of update live_condition_3:%v", err)
					return nil, err
				}
				_, err = stmtUpa.Exec(Condition.Price, Condition.Duration, Condition.TryToSee, Condition.VerificationCode,lid)
				if err != nil {
					log.Printf("Error of execution of update live_condition_3:%v", err)
					return nil,err
				}
				log.Printf("Update live_condition_3 success\n")

				emailList, err := RetrieveWhitelistByLid(lid)
				roomCondition := &defs.LiveRoomCondition{}
				roomCondition.ConditionType = 2
				roomCondition.Condition = 1
				roomCondition.WhiteUserList = emailList
				return roomCondition, nil
			}
		} else { //表示观看方式为验证码观看
			var newCode string
			if verificationCode == "" {
				code,_ := utils.NewStreamID()
				newCode = string(code[0:6])
			} else {
				newCode = verificationCode
			}
			stmtUpa, err := dbConn.Prepare("UPDATE live_condition SET lcondition = 1, condition_type = 3, price = ?, duration = ?, try_to_see = ?, verification_code = ? WHERE lid = ?")
			if err != nil {
				log.Printf("Error of preparation of update live_condition_4:%v", err)
				return nil, err
			}
			_, err = stmtUpa.Exec(Condition.Price, Condition.Duration, Condition.TryToSee, newCode, lid)
			if err != nil {
				log.Printf("Error of execution of update live_condition_4:%v", err)
				return nil,err
			}
			log.Printf("Update live_condition_4 success\n")
			roomCondition := &defs.LiveRoomCondition{}
			roomCondition.Condition = 1
			roomCondition.ConditionType = 3
			roomCondition.VerificationCode = newCode
			return roomCondition, nil
		}
	}
	return nil, nil
}

func RetrieveWhitelistByLid(lid string)(string, error) {
	stmtOut, err := dbConn.Prepare("SELECT email FROM whitelist WHERE lid = ?")
	if err != nil {
		log.Printf("Error of retrieve whitelist by lid:%v", err)
		return "", nil
	}
	rows, err := stmtOut.Query(lid)
	if err != nil {
		log.Printf("%s", err)
		return "", err
	}
	var email, emailList string
	for rows.Next() {
		if er := rows.Scan(&email); er != nil {
			log.Printf("Retrieve whitelist error: %s", er)
			return "", er
		}
		emailList += email + ";"
	}
	return emailList, nil
}
func RetrieveLRConditionByLid(lid string)(*defs.LiveRoomCondition, error) {
	stmtOut, err := dbConn.Prepare("SELECT lcondition, condition_type, price, duration, try_to_see, verification_code FROM live_condition WHERE lid = ?")
	if err != nil {
		log.Printf("%s", err)
		return nil, err
	}
	var condition, conditionType, tryToSee, duration int
	var price float32
	var verificationCode, emailList string
	err = stmtOut.QueryRow(lid).Scan(&condition, &conditionType, &price, &duration, &tryToSee, &verificationCode)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("%s", err)
		return nil, err
	}
	if conditionType == 2 && condition == 1 {
		emailList,_ = RetrieveWhitelistByLid(lid)
	}
	roomCondition := &defs.LiveRoomCondition{}
	roomCondition.Condition = condition
	roomCondition.ConditionType = conditionType
	roomCondition.Price = price
	roomCondition.Duration = duration
	roomCondition.TryToSee = tryToSee
	roomCondition.WhiteUserList = emailList
	roomCondition.VerificationCode = verificationCode
	roomCondition.Lid = lid

	return  roomCondition, nil

}

